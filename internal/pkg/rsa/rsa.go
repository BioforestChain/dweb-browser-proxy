package rsa

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	stringsHelper "github.com/BioforestChain/dweb-browser-proxy/internal/pkg/util/strings"
	"log"
)

// RSA公钥私钥产生
//fmt.Println("-------------------------------进行签名与验证操作-----------------------------------------")
//		fmt.Println("对消息进行签名操作...")
//		signData := RsaSignWithSha256([]byte(data), []byte(prvKey))
//		fmt.Println("消息的签名信息： ", hex.EncodeToString(signData))
//		fmt.Println("\n对签名信息进行验证...")
//		if RsaVerySignWithSha256([]byte(data), signData, []byte(pubKey)) {
//			fmt.Println("签名信息验证成功，确定是正确私钥签名！！")
//		}
//
//		fmt.Println("-------------------------------进行加密解密操作-----------------------------------------")
//		ciphertext := RsaEncrypt([]byte(data), []byte(pubKey))
//		fmt.Println("公钥加密后的数据：", hex.EncodeToString(ciphertext))
//		sourceData := RsaDecrypt(ciphertext, []byte(prvKey))
//		fmt.Println("私钥解密后的数据：", string(sourceData))

func GenRsaKey() (prvkey, pubkey []byte) {
	// 生成私钥文件
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		// TODO 日志上报
		log.Println("GenRsaKey privateKey panic: ", err)
	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	prvkey = pem.EncodeToMemory(block)
	publicKey := &privateKey.PublicKey
	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		// TODO 日志上报
		log.Println("GenRsaKey publicKey panic: ", err)
	}
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPkix,
	}
	pubkey = pem.EncodeToMemory(block)
	return
}

// 签名
func RsaSignWithSha256(data []byte, keyBytes []byte) []byte {
	h := sha256.New()
	h.Write(data)
	hashed := h.Sum(nil)
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		panic(errors.New("private key error"))
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		fmt.Println("ParsePKCS8PrivateKey err", err)
		panic(err)
	}

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed)
	if err != nil {
		fmt.Printf("Error from signing: %s\n", err)
		panic(err)
	}

	return signature
}

// RsaVerySignWithSha256
//
//	@Description: 验证填充方案使用RSASSA-PKCS1-v1_5而不是RSA-OAEP，这里通过传入的待验证数据，SPKI格式公钥，签名
//	@param data
//	@param signData
//	@param pubKeyPEM
//	@return bool
func RsaVerySignWithSha256(data, signData, pubKeyPEM string) bool {
	bytesSignData, err := base64.StdEncoding.DecodeString(signData)
	if err != nil {
		return false
	}
	bytesData := stringsHelper.StrToBytes(data)
	bytesPubKey := stringsHelper.StrToBytes(pubKeyPEM)

	// 解码秘钥
	block, _ := pem.Decode(bytesPubKey)
	if block == nil {
		fmt.Println("Failed to parse PEM block containing public key")
		return false
	}
	// 创建hash
	hash := crypto.SHA256.New()
	hash.Write(bytesData)
	digest := hash.Sum(nil)
	// 使用 SHA-256 哈希算法对消息进行哈希
	//hashed := sha256.Sum256(bytesData)
	//digest = hashed[:]
	// 解析SPKI格式公钥
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		fmt.Println("Error parsing public key:", err)
		return false
	}
	// 将公钥转换为 *rsa.PublicKey 类型
	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		fmt.Println("Not an RSA public key")
		return false
	}
	// 验证签名
	err = rsa.VerifyPKCS1v15(rsaPubKey, crypto.SHA256, digest, bytesSignData)
	if err != nil {
		fmt.Println("Signature verification failed:", err)
		return false
	}
	fmt.Println("Signature verified successfully!")
	return true
}

// 公钥加密
func RsaEncrypt(data, keyBytes []byte) []byte {
	//解密pem格式的公钥
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		panic(errors.New("public key error"))
	}
	// 解析公钥
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	// 类型断言
	pub := pubInterface.(*rsa.PublicKey)
	//加密
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, pub, data)
	if err != nil {
		panic(err)
	}
	return ciphertext
}

// 私钥解密
func RsaDecrypt(ciphertext, keyBytes []byte) []byte {
	//获取私钥
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		panic(errors.New("private key error!"))
	}
	//解析PKCS1格式的私钥
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	// 解密
	data, err := rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
	if err != nil {
		panic(err)
	}
	return data
}
