package sm4

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
)

// cbc mode
var DefaultInfo = []byte{'1', '1', 'H', 'D', 'E', 'S', 'a', 'A', 'h', 'i', 'H', 'H', 'u', 'g', 'D', 'z'}
var DefaultIv = []byte{'U', 'I', 'S', 'w', 'D', '9', 'f', 'W', '6', 'c', 'F', 'h', '9', 'S', 'N', 'S'}

// ecb mode
var DefaultEcb = []byte{'J', 'e', 'F', '8', 'U', '9', 'w', 'H', 'F', 'O', 'M', 'f', 's', '2', 'Y', '8'}

var sm4_FK = [4]uint32{0xa3b1bac6, 0x56aa3350, 0x677d9197, 0xb27022dc}

var sm4_CK = [32]uint32{
	0x00070e15, 0x1c232a31, 0x383f464d, 0x545b6269,
	0x70777e85, 0x8c939aa1, 0xa8afb6bd, 0xc4cbd2d9,
	0xe0e7eef5, 0xfc030a11, 0x181f262d, 0x343b4249,
	0x50575e65, 0x6c737a81, 0x888f969d, 0xa4abb2b9,
	0xc0c7ced5, 0xdce3eaf1, 0xf8ff060d, 0x141b2229,
	0x30373e45, 0x4c535a61, 0x686f767d, 0x848b9299,
	0xa0a7aeb5, 0xbcc3cad1, 0xd8dfe6ed, 0xf4fb0209,
	0x10171e25, 0x2c333a41, 0x484f565d, 0x646b7279,
}

var sm4_box = []byte{
	0xd6, 0x90, 0xe9, 0xfe, 0xcc, 0xe1, 0x3d, 0xb7, 0x16, 0xb6, 0x14, 0xc2, 0x28, 0xfb, 0x2c, 0x05,
	0x2b, 0x67, 0x9a, 0x76, 0x2a, 0xbe, 0x04, 0xc3, 0xaa, 0x44, 0x13, 0x26, 0x49, 0x86, 0x06, 0x99,
	0x9c, 0x42, 0x50, 0xf4, 0x91, 0xef, 0x98, 0x7a, 0x33, 0x54, 0x0b, 0x43, 0xed, 0xcf, 0xac, 0x62,
	0xe4, 0xb3, 0x1c, 0xa9, 0xc9, 0x08, 0xe8, 0x95, 0x80, 0xdf, 0x94, 0xfa, 0x75, 0x8f, 0x3f, 0xa6,
	0x47, 0x07, 0xa7, 0xfc, 0xf3, 0x73, 0x17, 0xba, 0x83, 0x59, 0x3c, 0x19, 0xe6, 0x85, 0x4f, 0xa8,
	0x68, 0x6b, 0x81, 0xb2, 0x71, 0x64, 0xda, 0x8b, 0xf8, 0xeb, 0x0f, 0x4b, 0x70, 0x56, 0x9d, 0x35,
	0x1e, 0x24, 0x0e, 0x5e, 0x63, 0x58, 0xd1, 0xa2, 0x25, 0x22, 0x7c, 0x3b, 0x01, 0x21, 0x78, 0x87,
	0xd4, 0x00, 0x46, 0x57, 0x9f, 0xd3, 0x27, 0x52, 0x4c, 0x36, 0x02, 0xe7, 0xa0, 0xc4, 0xc8, 0x9e,
	0xea, 0xbf, 0x8a, 0xd2, 0x40, 0xc7, 0x38, 0xb5, 0xa3, 0xf7, 0xf2, 0xce, 0xf9, 0x61, 0x15, 0xa1,
	0xe0, 0xae, 0x5d, 0xa4, 0x9b, 0x34, 0x1a, 0x55, 0xad, 0x93, 0x32, 0x30, 0xf5, 0x8c, 0xb1, 0xe3,
	0x1d, 0xf6, 0xe2, 0x2e, 0x82, 0x66, 0xca, 0x60, 0xc0, 0x29, 0x23, 0xab, 0x0d, 0x53, 0x4e, 0x6f,
	0xd5, 0xdb, 0x37, 0x45, 0xde, 0xfd, 0x8e, 0x2f, 0x03, 0xff, 0x6a, 0x72, 0x6d, 0x6c, 0x5b, 0x51,
	0x8d, 0x1b, 0xaf, 0x92, 0xbb, 0xdd, 0xbc, 0x7f, 0x11, 0xd9, 0x5c, 0x41, 0x1f, 0x10, 0x5a, 0xd8,
	0x0a, 0xc1, 0x31, 0x88, 0xa5, 0xcd, 0x7b, 0xbd, 0x2d, 0x74, 0xd0, 0x12, 0xb8, 0xe5, 0xb4, 0xb0,
	0x89, 0x69, 0x97, 0x4a, 0x0c, 0x96, 0x77, 0x7e, 0x65, 0xb9, 0xf1, 0x09, 0xc5, 0x6e, 0xc6, 0x84,
	0x18, 0xf0, 0x7d, 0xec, 0x3a, 0xdc, 0x4d, 0x20, 0x79, 0xee, 0x5f, 0x3e, 0xd7, 0xcb, 0x39, 0x48,
}

// 先base64解码，然后sm4解密
func DefaultDecodeBase64(inputStr string) ([]byte, error) {
	input, e := base64.StdEncoding.DecodeString(inputStr)
	if e != nil {
		return nil, e
	}
	return DefaultDecode(input)
}

// 先base64解码，然后sm4 ecb解密
func DecodeEcbBase64(inputStr string) ([]byte, error) {
	input, e := base64.StdEncoding.DecodeString(inputStr)
	if e != nil {
		return nil, e
	}
	return Decode(input, DefaultEcb)
}

func DefaultEncode(input []byte) ([]byte, error) {
	a := string(DefaultInfo)
	b := string(DefaultIv)
	return EncodeCBC(input, []byte(a), []byte(b))
}

func DefaultDecode(input []byte) ([]byte, error) {
	a := string(DefaultInfo)
	b := string(DefaultIv)
	return DecodeCBC(input, []byte(a), []byte(b))
}

func EncodeECB(input, key []byte) ([]byte, error) {
	if len(key) != 16 {
		return nil, errors.New("secret key length is not 16")
	}
	//1.补齐16位，如果正好是16的倍数，也补齐操作
	inputLen := len(input)
	paddingLen := 16 - inputLen%16
	for i := 0; i < paddingLen; i++ {
		input = append(input, byte(paddingLen))
	}

	//2.生成加密轮秘钥
	//2.1 加密密钥MK=(MK 0 , MK 1 , MK 2 , MK 3 )，MK i ∈ ，i＝0,1,2,3
	MK := make([]uint32, 4)
	MK[0] = binary.BigEndian.Uint32(key[0:4])
	MK[1] = binary.BigEndian.Uint32(key[4:8])
	MK[2] = binary.BigEndian.Uint32(key[8:12])
	MK[3] = binary.BigEndian.Uint32(key[12:16])
	//2.2 (K 0 ,K 1 ,K 2 ,K 3 )=(MK 0 ○+ FK 0 ,MK 1 ○+ FK 1 ,MK 2 ○+ FK 2 ,MK 3 ○+ FK 3 )
	K := make([]uint32, 36)
	for i := 0; i < 4; i++ {
		K[i] = MK[i] ^ sm4_FK[i]
	}
	//2.3 K[4:35]就是轮秘钥rk[0:31]
	for i := 0; i < 32; i++ {
		K[i+4] = K[i] ^ sm4_fun_L(K[i+1]^K[i+2]^K[i+3]^sm4_CK[i])
	}
	rk := K[4:]

	//3加密过程
	X := make([]uint32, 36)
	outputLen := len(input)
	output := make([]byte, outputLen)
	for i := 0; i < outputLen; i += 16 {
		//3.1赋值
		X[0] = binary.BigEndian.Uint32(input[i : i+4])
		X[1] = binary.BigEndian.Uint32(input[i+4 : i+8])
		X[2] = binary.BigEndian.Uint32(input[i+8 : i+12])
		X[3] = binary.BigEndian.Uint32(input[i+12 : i+16])

		for j := 0; j < 32; j++ {
			X[j+4] = X[j] ^ sm4_fun_T(X[j+1]^X[j+2]^X[j+3]^rk[j])
		}

		binary.BigEndian.PutUint32(output[i:i+4], X[35])
		binary.BigEndian.PutUint32(output[i+4:i+8], X[34])
		binary.BigEndian.PutUint32(output[i+8:i+12], X[33])
		binary.BigEndian.PutUint32(output[i+12:i+16], X[32])
	}

	return output, nil
}

func Decode(input, key []byte) ([]byte, error) {
	if len(key) != 16 {
		return nil, errors.New("secret key length is not 16")
	}

	//1.生成加密轮秘钥
	//1.1 加密密钥MK=(MK 0 , MK 1 , MK 2 , MK 3 )，MK i ∈ ，i＝0,1,2,3
	MK := make([]uint32, 4)
	MK[0] = binary.BigEndian.Uint32(key[0:4])
	MK[1] = binary.BigEndian.Uint32(key[4:8])
	MK[2] = binary.BigEndian.Uint32(key[8:12])
	MK[3] = binary.BigEndian.Uint32(key[12:16])
	//1.2 (K 0 ,K 1 ,K 2 ,K 3 )=(MK 0 ○+ FK 0 ,MK 1 ○+ FK 1 ,MK 2 ○+ FK 2 ,MK 3 ○+ FK 3 )
	K := make([]uint32, 36)
	for i := 0; i < 4; i++ {
		K[i] = MK[i] ^ sm4_FK[i]
	}
	//1.3 K[4:35]就是轮秘钥rk[0:31]
	for i := 0; i < 32; i++ {
		K[i+4] = K[i] ^ sm4_fun_L(K[i+1]^K[i+2]^K[i+3]^sm4_CK[i])
	}
	rk := K[4:]
	for i := 0; i < 16; i++ {
		rk[i], rk[31-i] = rk[31-i], rk[i]
	}

	//2加密过程
	X := make([]uint32, 36)
	outputLen := len(input)
	if outputLen%16 != 0 {
		return nil, errors.New("decrpt error")
	}
	output := make([]byte, outputLen)
	for i := 0; i < outputLen; i += 16 {
		//2.1赋值
		X[0] = binary.BigEndian.Uint32(input[i : i+4])
		X[1] = binary.BigEndian.Uint32(input[i+4 : i+8])
		X[2] = binary.BigEndian.Uint32(input[i+8 : i+12])
		X[3] = binary.BigEndian.Uint32(input[i+12 : i+16])

		for j := 0; j < 32; j++ {
			X[j+4] = X[j] ^ sm4_fun_T(X[j+1]^X[j+2]^X[j+3]^rk[j])
		}

		binary.BigEndian.PutUint32(output[i:i+4], X[35])
		binary.BigEndian.PutUint32(output[i+4:i+8], X[34])
		binary.BigEndian.PutUint32(output[i+8:i+12], X[33])
		binary.BigEndian.PutUint32(output[i+12:i+16], X[32])
	}

	//3.补齐16位，如果正好是16的倍数，也补齐操作
	paddingLen := int(output[outputLen-1])
	if outputLen < paddingLen {
		return nil, errors.New("decrpt error")
	}
	return output[0 : outputLen-paddingLen], nil
}

func EncodeCBC(input, key, iv []byte) ([]byte, error) {
	if len(input) == 0 {
		return nil, errors.New("input is nil")
	}

	if len(key) != 16 {
		return nil, errors.New("secret key length is not 16")
	}

	if len(iv) != 16 {
		return nil, errors.New("secret iv length is not 16")
	}
	//1.补齐16位，如果正好是16的倍数，也补齐操作
	inputLen := len(input)
	paddingLen := 16 - inputLen%16
	for i := 0; i < paddingLen; i++ {
		input = append(input, byte(paddingLen))
	}

	//2.生成加密轮秘钥
	//2.1 加密密钥MK=(MK 0 , MK 1 , MK 2 , MK 3 )，MK i ∈ ，i＝0,1,2,3
	MK := make([]uint32, 4)
	MK[0] = binary.BigEndian.Uint32(key[0:4])
	MK[1] = binary.BigEndian.Uint32(key[4:8])
	MK[2] = binary.BigEndian.Uint32(key[8:12])
	MK[3] = binary.BigEndian.Uint32(key[12:16])
	//2.2 (K 0 ,K 1 ,K 2 ,K 3 )=(MK 0 ○+ FK 0 ,MK 1 ○+ FK 1 ,MK 2 ○+ FK 2 ,MK 3 ○+ FK 3 )
	K := make([]uint32, 36)
	for i := 0; i < 4; i++ {
		K[i] = MK[i] ^ sm4_FK[i]
	}
	//2.3 K[4:35]就是轮秘钥rk[0:31]
	for i := 0; i < 32; i++ {
		K[i+4] = K[i] ^ sm4_fun_L(K[i+1]^K[i+2]^K[i+3]^sm4_CK[i])
	}
	rk := K[4:]

	//3加密过程
	X := make([]uint32, 36)
	outputLen := len(input)
	output := make([]byte, outputLen)
	for i := 0; i < outputLen; i += 16 {
		//3.1 iv
		for j := 0; j < 16; j++ {
			output[i+j] = input[i+j] ^ iv[j]
		}
		//3.2赋值
		X[0] = binary.BigEndian.Uint32(output[i : i+4])
		X[1] = binary.BigEndian.Uint32(output[i+4 : i+8])
		X[2] = binary.BigEndian.Uint32(output[i+8 : i+12])
		X[3] = binary.BigEndian.Uint32(output[i+12 : i+16])

		for j := 0; j < 32; j++ {
			X[j+4] = X[j] ^ sm4_fun_T(X[j+1]^X[j+2]^X[j+3]^rk[j])
		}

		binary.BigEndian.PutUint32(output[i:i+4], X[35])
		binary.BigEndian.PutUint32(output[i+4:i+8], X[34])
		binary.BigEndian.PutUint32(output[i+8:i+12], X[33])
		binary.BigEndian.PutUint32(output[i+12:i+16], X[32])

		//3.3 iv
		for j := 0; j < 16; j++ {
			iv[j] = output[i+j]
		}
	}

	return output, nil
}

func DecodeCBC(input, key, iv []byte) ([]byte, error) {
	if len(input) == 0 {
		return nil, errors.New("input is nil")
	}

	if len(key) != 16 {
		return nil, errors.New("secret key length is not 16")
	}

	if len(iv) != 16 {
		return nil, errors.New("secret iv length is not 16")
	}

	//1.生成加密轮秘钥
	//1.1 加密密钥MK=(MK 0 , MK 1 , MK 2 , MK 3 )，MK i ∈ ，i＝0,1,2,3
	MK := make([]uint32, 4)
	MK[0] = binary.BigEndian.Uint32(key[0:4])
	MK[1] = binary.BigEndian.Uint32(key[4:8])
	MK[2] = binary.BigEndian.Uint32(key[8:12])
	MK[3] = binary.BigEndian.Uint32(key[12:16])
	//1.2 (K 0 ,K 1 ,K 2 ,K 3 )=(MK 0 ○+ FK 0 ,MK 1 ○+ FK 1 ,MK 2 ○+ FK 2 ,MK 3 ○+ FK 3 )
	K := make([]uint32, 36)
	for i := 0; i < 4; i++ {
		K[i] = MK[i] ^ sm4_FK[i]
	}
	//1.3 K[4:35]就是轮秘钥rk[0:31]
	for i := 0; i < 32; i++ {
		K[i+4] = K[i] ^ sm4_fun_L(K[i+1]^K[i+2]^K[i+3]^sm4_CK[i])
	}
	rk := K[4:]
	for i := 0; i < 16; i++ {
		rk[i], rk[31-i] = rk[31-i], rk[i]
	}

	//2加密过程
	X := make([]uint32, 36)
	outputLen := len(input)
	if outputLen%16 != 0 {
		return nil, errors.New("decrpt error")
	}
	output := make([]byte, outputLen)
	for i := 0; i < outputLen; i += 16 {
		//2.1赋值
		X[0] = binary.BigEndian.Uint32(input[i : i+4])
		X[1] = binary.BigEndian.Uint32(input[i+4 : i+8])
		X[2] = binary.BigEndian.Uint32(input[i+8 : i+12])
		X[3] = binary.BigEndian.Uint32(input[i+12 : i+16])

		for j := 0; j < 32; j++ {
			X[j+4] = X[j] ^ sm4_fun_T(X[j+1]^X[j+2]^X[j+3]^rk[j])
		}

		binary.BigEndian.PutUint32(output[i:i+4], X[35])
		binary.BigEndian.PutUint32(output[i+4:i+8], X[34])
		binary.BigEndian.PutUint32(output[i+8:i+12], X[33])
		binary.BigEndian.PutUint32(output[i+12:i+16], X[32])

		//3.2
		for j := 0; j < 16; j++ {
			output[i+j] = output[i+j] ^ iv[j]
		}

		//3.3 iv
		for j := 0; j < 16; j++ {
			iv[j] = input[i+j]
		}
	}

	//3.补齐16位，如果正好是16的倍数，也补齐操作
	paddingLen := int(output[outputLen-1])
	if outputLen < paddingLen {
		return nil, errors.New("decrpt error")
	}
	return output[0 : outputLen-paddingLen], nil
}

func sm4_fun_L(u uint32) uint32 {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, u)
	sb := make([]byte, 4)
	sb[0] = sm4_box[b[0]]
	sb[1] = sm4_box[b[1]]
	sb[2] = sm4_box[b[2]]
	sb[3] = sm4_box[b[3]]

	m := binary.BigEndian.Uint32(sb)
	return (m ^ sm4_fun_left(m, 13) ^ sm4_fun_left(m, 23))
}

func sm4_fun_T(u uint32) uint32 {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, u)
	sb := make([]byte, 4)
	sb[0] = sm4_box[b[0]]
	sb[1] = sm4_box[b[1]]
	sb[2] = sm4_box[b[2]]
	sb[3] = sm4_box[b[3]]

	m := binary.BigEndian.Uint32(sb)

	return (m ^ sm4_fun_left(m, 2) ^ sm4_fun_left(m, 10) ^ sm4_fun_left(m, 18) ^ sm4_fun_left(m, 24))
}

func sm4_fun_left(u uint32, n uint32) uint32 {
	//将无符号整形循环左移n位 <<<
	left := n % 32
	return ((u << left) | (u >> (32 - left)))
}
