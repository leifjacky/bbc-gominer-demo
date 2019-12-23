package main

/*
void cryptonightbbcslow(char *input, int size, char *output, int variant, int prehashed);
#cgo LDFLAGS: -L. -Llib  -lbbc
*/
import "C"
import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"github.com/sirupsen/logrus"
)

const (
	RESPONSE_DEFAULT int = iota
	RESPONSE_LOGIN
)

func BBCHashSlow(data []byte) []byte {
	output := make([]byte, 32)
	C.cryptonightbbcslow((*C.char)(unsafe.Pointer(&data[0])), C.int(len(data)), (*C.char)(unsafe.Pointer(&output[0])), 2, 0)
	return output
}

type StratumMiner struct {
	cfg *StratumMinerConfig

	job atomic.Value
	cnt int64

	writeMu sync.Mutex
	conn    net.Conn
}

type Job struct {
	sync.Mutex
	jobId         string
	blob          string
	stratumTarget string
	target        *big.Int
	ntime         string
	nonce         uint64
}

func (j *Job) GetNextNonce() string {
	j.Lock()
	defer j.Unlock()
	n := j.nonce
	if j.nonce == 0xffffffffffffffff {
		j.nonce &= 0xffffffff00000000
	} else {
		j.nonce++
	}
	return ReverseStringByte(fmt.Sprintf("%016x", n))
}

func NewMiner(cfg *StratumMinerConfig) *StratumMiner {
	return &StratumMiner{
		cfg: cfg,
	}
}

func (m *StratumMiner) Mine() {
	gracefulShutdownChannel := make(chan os.Signal)
	signal.Notify(gracefulShutdownChannel, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-gracefulShutdownChannel
		logrus.Warningf("receive shutdown signal")
		os.Exit(0)
	}()

	sumIntv := MustParseDuration(m.cfg.SumIntv)
	logrus.Infof("hashrate sum every %v", sumIntv)
	sumTicker := time.NewTicker(sumIntv)

	go m.start()
	for {
		select {
		case <-sumTicker.C:
			cnt := m.cnt
			m.cnt -= cnt
			logrus.Warningf("hashrates: %v", GetReadableHashRateString(float64(cnt/int64((sumIntv)/time.Second))))
		}
	}
}

func (m *StratumMiner) start() {
	th := m.cfg.Threads
	if th == 0 {
		th = runtime.NumCPU()
	}
	logrus.Infof("running with %v workers", th)
	for i := 0; i < th; i++ {
		go m.startWorker(i)
	}

	logrus.Infof("connect to %v", m.cfg.Url)
	conn, err := net.Dial("tcp", m.cfg.Url)
	if err != nil {
		logrus.Fatalf("failed to connect: %v", err)
	}
	m.conn = conn
	logrus.Infof("connected")

	buf := bufio.NewReader(conn)

	if err := m.request("login", map[string]string{
		"login": m.cfg.Username,
		"pass":  m.cfg.Password,
		"agent": "bbc gominer demo 1.0.0",
	}); err != nil {
		logrus.Fatalf("error authorize: %v", err)
	}
	data, _, err := buf.ReadLine()
	if err != nil {
		logrus.Errorf("err reading: %v", err)
		return
	}
	logrus.Debugf("recv from pool: %v", string(data))
	if err := m.handleMesg(data, RESPONSE_LOGIN); err != nil {
		logrus.Errorf("err handle mesg: %v", err)
		return
	}
	logrus.Infof("authorized")

	for {
		data, _, err := buf.ReadLine()
		if err != nil {
			logrus.Errorf("err reading: %v", err)
			return
		}

		logrus.Debugf("recv from pool: %v", string(data))
		if err := m.handleMesg(data, RESPONSE_DEFAULT); err != nil {
			logrus.Errorf("err handle mesg: %v", err)
			return
		}
	}
	logrus.Infof("disconnected")
}

func (m *StratumMiner) handleMesg(line []byte, flag int) error {
	var mesg PoolMesg
	if err := json.Unmarshal(line, &mesg); err != nil {
		return fmt.Errorf("can't decode: %v", err)
	}
	switch flag {
	case RESPONSE_LOGIN:
		if mesg.Error != nil {
			return fmt.Errorf("login error. %v(%d)", mesg.Error.Message, mesg.Error.Code)
		}
		job := JsonResult{}
		if err := json.Unmarshal(*mesg.Result, &job); err != nil {
			return fmt.Errorf("can't decode params: %v", err)
		}
		m.handleJob(&job.Job)
		return nil
	}
	switch mesg.Method {
	case "job":
		job := JsonJob{}
		if err := json.Unmarshal(*mesg.Params, &job); err != nil {
			return fmt.Errorf("can't decode params: %v", err)
		}
		m.handleJob(&job)
	default:
		result := JsonResult{}
		if err := json.Unmarshal(*mesg.Result, &result); err != nil {
			if mesg.Error == nil {
				logrus.Infof("share rejected.")
			} else {
				logrus.Infof("share rejected. %v(%d)", mesg.Error.Message, mesg.Error.Code)
			}
		}
		if result.Status == "OK" {
			logrus.Infof("share accepted.")
		}
	}
	return nil
}

func (m *StratumMiner) handleJob(job *JsonJob) {
	newJob := Job{
		jobId:         job.JobId,
		blob:          job.Blob,
		stratumTarget: job.Target,
		target:        BigbangStratumTargetStr2BigTarget(job.Target),
		ntime:         job.Blob[8:16],
		nonce:         MustParseUInt64(ReverseStringByte(job.Blob[218:234]), 16),
	}
	logrus.Infof("new job: %v - %v %064x", newJob.jobId, newJob.blob, newJob.target)
	m.job.Store(&newJob)
}

type JsonRpcReq struct {
	Id     int64       `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

type PoolMesg struct {
	Id     *json.RawMessage `json:"id"`
	Method string           `json:"method"`
	Result *json.RawMessage `json:"result"`
	Params *json.RawMessage `json:"params"`
	Error  *JsonError       `json:"error"`
}

type JsonResult struct {
	Status string  `json:"status"`
	Id     string  `json:"id"`
	Job    JsonJob `json:"job"`
}

type JsonJob struct {
	Blob   string `json:"blob"`
	Id     string `json:"id"`
	JobId  string `json:"job_id"`
	Target string `json:"target"`
}

type JsonError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (m *StratumMiner) request(method string, params interface{}) error {
	return m.write(&JsonRpcReq{0, method, params})
}

var lineDelimiter = []byte("\n")

func (m *StratumMiner) write(message interface{}) error {
	b, err := json.Marshal(message)
	if err != nil {
		return err
	}

	m.writeMu.Lock()
	defer m.writeMu.Unlock()

	logrus.Debugf("write to pool: %v", string(b))
	if _, err := m.conn.Write(b); err != nil {
		return err
	}

	_, err = m.conn.Write(lineDelimiter)
	return err
}

func (m *StratumMiner) loadJob() *Job {
	job := m.job.Load()
	if job == nil {
		return nil
	}
	return job.(*Job)
}

func (m *StratumMiner) startWorker(i int) {
	for {
		job := m.loadJob()
		if job == nil {
			logrus.Warningf("#%d job not ready. sleep for 5s...", i)
			time.Sleep(5 * time.Second)
			continue
		}
		blob := job.blob
		nonce := job.GetNextNonce()
		b := MustStringToHexBytes(blob[:218] + nonce)
		hash := BBCHashSlow(b)
		bInt := Hash2BigTarget(ReverseBytes(hash))
		if bInt.Cmp(job.target) <= 0 {
			logrus.Tracef("solve %x %064x", b, hash)
			logrus.Infof("share found: %s - %064x", nonce[:8], bInt)
			go func() {
				if err := m.request("submit", map[string]string{
					"id":     job.jobId,
					"job_id": job.jobId,
					"nonce":  nonce[:8],
					"time":   job.ntime,
					"result": ReverseStringByte(fmt.Sprintf("%064x", bInt)),
				}); err != nil {
					logrus.Fatalf("error submit: %v", err)
				}
			}()
		}
		m.cnt++
	}
}
