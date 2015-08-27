package base_test

import (
	"crypto/aes"
	"testing"

	"github.com/itpkg/base"
)

const hello = "Hello, IT-PACKAGE!!!"

var key = []byte("11111111111111111111111111111111")

var cip, _ = aes.NewCipher(key)
var helper = base.Helper{Key: key, Cip: cip}

func TestRandom(t *testing.T) {
	t.Logf("Random string: %s\t%s", helper.RandomStr(16), helper.RandomStr(16))
}

func TestHmac(t *testing.T) {

	dest1 := helper.HmacSum([]byte(hello))
	dest2 := helper.HmacSum([]byte(hello))

	t.Logf("HMAC1(%d): %x", len(dest1), dest1)
	t.Logf("HMAC2(%d): %x", len(dest2), dest2)
	if !helper.HmacEqual(dest1, dest2) {
		t.Errorf("HMAC FAILED!")
	}
}

func TestMd5AndSha(t *testing.T) {
	t.Logf("MD5: %x", helper.Md5([]byte(hello)))
}

func TestBase64(t *testing.T) {
	dest := helper.Base64Encode([]byte(hello))
	t.Logf("Base64: %s => %x", hello, dest)
	src, err := helper.Base64Decode(dest)
	if err != nil || string(src) != hello {
		t.Errorf("val == %x, want %x", src, hello)
	}
}

func TestAes(t *testing.T) {

	dest1, iv1, _ := helper.AesEncrypt([]byte(hello))
	dest2, iv2, _ := helper.AesEncrypt([]byte(hello))
	t.Logf("AES1(%d, iv=%x): %s => %x", len(dest1), iv1, hello, dest1)
	t.Logf("AES2(%d, iv=%x): %s => %x", len(dest2), iv2, hello, dest2)

	src := helper.AesDecrypt(dest1, iv1)
	if string(src) != hello {
		t.Errorf("val == %x, want %x", src, hello)
	}
}