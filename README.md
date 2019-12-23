## Bigbang(BBC) cpuminer written in Golang

### usage

```bash
cd bbc
mkdir build && cd build
cmake .. && make
cd ../..
ln -fs bbc/build/lib .
LD_LIBRARY_PATH=lib go run *.go
```

## bbc mining real world example

### login

```json
{
	"id": 1,
	"jsonrpc": "2.0",
	"method": "login",
	"params": {
		"login": "1rmycpneemett21ejf6vnan0twktdqah5scqd2ybrkhfyf2x9yr038r44.workername",
		"pass": "x",
		"agent": "XMRig/3.1.3 (Linux x86_64) libuv/1.8.0 gcc/7.4.0",
		"rigid": "workername" // rig_id填写workername，或者直接加在地址后
	}
}

{
	"jsonrpc": "2.0",
	"id": 1,
	"result": {
		"id": "85396353",
		"job": {
      "blob": "01000100c9c0e85d477b18fa29ff157929e9de4df986873880fe341ae72adb50763e85a55a2100004c0000000000000000000000000000000000000000000000000000000000000000000127020400ff0b69caf0cf371d2b28b7fb4cd34887bb38f944090038e7b57dc82fec9d000000007a331cb4",
      "id": "85392734",
      "job_id": "85392734",
      "target": "ffffffffffff0100"
		},
		"status": "OK"
	},
	"error": null
}
```



### job notify

```json
{
	"jsonrpc": "2.0",
	"method": "job",
	"params": {
		"blob": "01000100c9c0e85d477b18fa29ff157929e9de4df986873880fe341ae72adb50763e85a55a2100004c0000000000000000000000000000000000000000000000000000000000000000000127020400ff0b69caf0cf371d2b28b7fb4cd34887bb38f944090038e7b57dc82fec9d000000007a331cb4",
		"id": "85392734",
		"job_id": "85392734",
		"target": "ffffffffffff0100"
	}
}
```

协议中target与diff的转换：

targetLimit = 1 <<  64 - 1

target = reverseBytes( targetLimit >> diff )



### submit

```json
{
	"id": 3,
	"jsonrpc": "2.0",
	"method": "submit",
	"params": {
		"id": "85396353",
		"job_id": "85391651",
		"nonce": "b2250000",
		"time": "e2c0e85d",
		"result": "c06882dbd3b704fa3a1c77bcd3024c9d2750acb8afa48e94fb3d54c9e02e0100"
	}
}

accept:
{
	"jsonrpc": "2.0",
	"id": 3,
	"result": {
		"status": "OK"
	},
	"error": null
}

reject:
{
	"jsonrpc": "2.0",
	"id": 3,
	"result": false
	"error": {
		"code": 23,
		"message": "duplicate share"
	}
}
```



```json
In this example

blob = 01000100c9c0e85d477b18fa29ff157929e9de4df986873880fe341ae72adb50763e85a55a2100004c0000000000000000000000000000000000000000000000000000000000000000000127020400ff0b69caf0cf371d2b28b7fb4cd34887bb38f944090038e7b57dc82fec9d000000007a331cb4 (length 234)
pow = blob[:8] + ntime + blob[16:218] + nonce + blob[226:] = 01000100e2c0e85d477b18fa29ff157929e9de4df986873880fe341ae72adb50763e85a55a2100004c0000000000000000000000000000000000000000000000000000000000000000000127020400ff0b69caf0cf371d2b28b7fb4cd34887bb38f944090038e7b57dc82fec9db22500007a331cb4
reverseHash = reverseBytes(BBCHash(pow)) = c06882dbd3b704fa3a1c77bcd3024c9d2750acb8afa48e94fb3d54c9e02e0100
hash = 00012ee0c9543dfb948ea4afb8ac50279d4c02d3bc771c3afa04b7d3db8268c0

jobTargetReverse = ffffffffffff0100
jobTarget = 0001ffffffffffff
hashTarget = 0001ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff

compare hash and hashTarget
```


