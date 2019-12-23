CMAKE_MINIMUM_REQUIRED(VERSION 2.8.4)

SET(LIBBBC_SRC
	./cryptonight/aesb.c
	./cryptonight/blake256.c
	./cryptonight/groestl.c
	./cryptonight/hash.c
	./cryptonight/hash-extra-blake.c
	./cryptonight/hash-extra-groestl.c
	./cryptonight/hash-extra-jh.c
	./cryptonight/hash-extra-skein.c
	./cryptonight/jh.c
	./cryptonight/keccak.c
	./cryptonight/oaes_lib.c
	./cryptonight/skein.c
	./cryptonight/slow-hash.c
	./cryptonight.c
	)
SET(CMAKE_C_FLAGS "-Wall -O3 -march=native -mtune=generic")

INCLUDE_DIRECTORIES(.)
INCLUDE_DIRECTORIES(./cryptonight)

ADD_LIBRARY(bbc SHARED ${LIBBBC_SRC})

#SET_TARGET_PROPERTIES(cryptonight PROPERTIES OUTPUT_NAME "cryptonight")

SET(LIBRARY_OUTPUT_PATH ${PROJECT_BINARY_DIR}/lib)

INSTALL(TARGETS bbc
	LIBRARY DESTINATION lib
	ARCHIVE DESTINATION lib)
