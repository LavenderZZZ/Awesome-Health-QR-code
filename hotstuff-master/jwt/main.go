package jwt

import (
	"errors"
	"time"
)

var secret="4680bbfa2dc33ab4f5d657658156a075"


type User struct {
	Userid   string
	Username string
	Passwd string
	Status string
	Picday int
}


type userclaims struct {
	jwt.StandardClaims
	*User
}

//颁发token
func jwtGenerateToken(m *User,d time.Duration)(string,error){
	m.Passwd=""
	expiretime:=time.Now().Add(d)
	stdClaims := jwt.StandardClaims{
		ExpiresAt: expireTime.Unix(),
		IssuedAt:  time.Now().Unix(),
		Id:        m.Userid,
		Issuer:    "127.0.0.1",
	}
	uClaims:=userclaims{
		stdClaims,
		m,
	}

	token,err:=jwt.NewWithClaims(jwt.SigningMethodHS256, uClaims).SignedString([]byte(secret))
	return token,err
}

//解析token
func JwtParserUser(token string)(*User,int64,error){
	if token==""{
		return nil,0,errors.New("Token is not null")
	}

	claims:=userclaims{}
	_,err:=jwt.NewWithClaims(token,&claims,func(t *jwt)(interface{},
	return []byte(secret),nil})

}