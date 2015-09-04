package base

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	c_rand "crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"log/syslog"
	m_rand "math/rand"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/garyburd/redigo/redis"
	"github.com/jrallison/go-workers"
)

type Helper struct {
	HKey []byte `inject:"hmac.key"`
	//16、24或者32位的[]byte，分别对应AES-128, AES-192或AES-256算法
	Cip    cipher.Block   `inject:"aes.cip"`
	Redis  *redis.Pool    `inject:""`
	Logger *syslog.Writer `inject:""`
}

func (p *Helper) TokenParse(ticket string) (map[string]interface{}, error) {
	if data, err := jwt.Parse(ticket, func(token *jwt.Token) (interface{}, error) {
		c := p.Redis.Get()
		defer c.Close()
		return redis.Bytes(c.Do("GET", p.TokenId(token.Header["kid"].(string))))
	}); err == nil {
		if data.Valid {
			return data.Claims["data"].(map[string]interface{}), nil
		} else {
			return nil, errors.New("error.bad_token")
		}

	} else {
		fmt.Printf("### %v: [%s]\n", err, ticket)
		return nil, err
	}
}

func (p *Helper) TokenId(kid string) string {
	return fmt.Sprintf("token://%s", kid)
}

func (p *Helper) TokenTtl(kid string, minutes int) error {
	c := p.Redis.Get()
	defer c.Close()
	var err error
	if minutes > 0 {
		_, err = c.Do("EXPIRE", p.TokenId(kid), 60*minutes)
	} else {
		_, err = c.Do("DEL", p.TokenId(kid))
	}

	return err
}

func (p *Helper) TokenCreate(kid string, data map[string]interface{}, minutes int) (string, error) {
	c := p.Redis.Get()
	defer c.Close()

	key, err := redis.Bytes(c.Do("GET", p.TokenId(kid)))
	if err != nil {
		key, err = p.RandomBytes(32)
		if err != nil {
			return "", err
		}
		if minutes > 0 {
			_, err = c.Do("SET", p.TokenId(kid), key, "EX", 60*minutes)
		} else {
			_, err = c.Do("SET", p.TokenId(kid), key)
		}
	}

	if err != nil {
		return "", err
	}

	token := jwt.New(jwt.SigningMethodHS512)
	token.Header["kid"] = kid
	token.Claims["data"] = data
	token.Claims["exp"] = time.Now().Add(time.Duration(minutes) * time.Minute).Unix()
	return token.SignedString(key)

}

func (p *Helper) Shell(cmd string, args ...string) error {
	bin, err := exec.LookPath(cmd)
	if err != nil {
		return err
	}
	return syscall.Exec(bin, append([]string{cmd}, args...), os.Environ())
}

func (p *Helper) RandomStr(size int) string {
	const letters = "0123456789abcdefghijklmnopqrstuvwxyz"
	buf := make([]byte, size)
	for i := range buf {
		buf[i] = letters[m_rand.Intn(len(letters))]
	}
	return string(buf)

}

func (p *Helper) RandomBytes(size int) ([]byte, error) {
	b := make([]byte, size)
	if _, err := c_rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

func (p *Helper) Md5(src []byte) [16]byte {
	return md5.Sum(src)
}

func (p *Helper) Base64Encode(src []byte) []byte {
	return []byte(base64.StdEncoding.EncodeToString(src))
}

func (p *Helper) Base64Decode(src []byte) ([]byte, error) {
	return base64.StdEncoding.DecodeString(string(src))
}

func (p *Helper) HmacSum(src []byte) []byte {
	mac := hmac.New(sha512.New, p.HKey)
	mac.Write(src)
	return mac.Sum(nil)
}

func (p *Helper) HmacEqual(src, dst []byte) bool {
	return hmac.Equal(p.HmacSum(src), dst)
}

func (p *Helper) AesEncrypt(src []byte) ([]byte, []byte, error) {
	iv, err := p.RandomBytes(aes.BlockSize)
	if err != nil {
		return nil, nil, err
	}
	cfb := cipher.NewCFBEncrypter(p.Cip, iv)
	ct := make([]byte, len(src))
	cfb.XORKeyStream(ct, src)
	return ct, iv, nil

}

func (p *Helper) AesDecrypt(src, iv []byte) []byte {
	cfb := cipher.NewCFBDecrypter(p.Cip, iv)
	pt := make([]byte, len(src))
	cfb.XORKeyStream(pt, src)
	return pt
}

func (p *Helper) Obj2bits(obj interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(obj)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p *Helper) Bits2obj(data []byte, obj interface{}) error {
	var buf bytes.Buffer
	dec := gob.NewDecoder(&buf)
	buf.Write(data)
	err := dec.Decode(obj)
	if err != nil {
		return err
	}
	return nil
}

func (p *Helper) Obj2json(o interface{}) (string, error) {
	b, e := json.Marshal(o)
	return string(b), e
}

func (p *Helper) Json2obj(j string, o interface{}) error {
	return json.Unmarshal([]byte(j), o)
}

func (p *Helper) BackJob(queue, class string, args interface{}) {
	workers.Enqueue(queue, class, args)
}
