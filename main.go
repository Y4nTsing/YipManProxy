package main

import (
	"flag"
	"fmt"
	socks5 "github.com/armon/go-socks5"
	"github.com/hashicorp/yamux"
	"io"
	"net"
	"log"
	"math/rand"
	"time"
)

var clientServerConns []*yamux.Session
var encKey [32]byte

func listenClientServerConn(port string){
	log.Println("Listening for the ClientServer")
	ln, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		panic(err)
	}
	for{
		conn, err := ln.Accept()
		if err != nil {
			log.Println("ClientServer connect fail")
		}
		buf := make([]byte, 32)
		conn.Read(buf)
		if string(buf)==string(encKey[:]) {
			session, err := yamux.Client(conn, nil)
			if err != nil {
				log.Println("ClientServer session create failed")
			}
			clientServerConns = append(clientServerConns, session)
		}
	}
}

func listenClientClientConn(port string){
	log.Println("Listening for the ClientClient")
	ln, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		panic(err)
	}
	for{
		conn, err := ln.Accept()
		if err != nil {
			log.Println("ClientClient connect fail")
		}

		if len(clientServerConns)>0 {
			//流量转发到客户端服务器
			clientServerConn:=getStream()

			forward :=func(src,dest net.Conn){
				defer src.Close()
				defer dest.Close()
				io.Copy(dest,src)
			}

			go forward(clientServerConn,conn)
			go forward(conn,clientServerConn)
		}else{
			log.Print("there is no client server...")
			time.Sleep(5 * time.Second)
		}

	}
}

func clientServerServeSocks(remoteAddress,username,password string){
	//设置sock5账号密码
	cred := socks5.StaticCredentials{
		username: password,
	}
	cator := socks5.UserPassAuthenticator{Credentials: cred}
	conf := &socks5.Config{AuthMethods:[]socks5.Authenticator{cator}}
	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}

	log.Println("Connecting to server...")
	for {
		conn := func(remoteAddress string) (conn net.Conn) {
			for {
				conn, err := net.Dial("tcp", remoteAddress)
				if err != nil {
					log.Print("trying to reconnect to server...")
					time.Sleep(5 * time.Second)
					continue
				}
				conn.Write(encKey[:])
				return conn
			}
		}(remoteAddress)
		session, err := yamux.Server(conn, nil)
		if err != nil {
			continue
		}
		for {
			stream, err := session.Accept()
			if err != nil {
				break
			}
			go func() {
				err = server.ServeConn(stream)
				if err != nil {
					log.Println(err)
				}
			}()
		}
	}

}

func getStream() net.Conn{
	for {
		clientServerConnsLen:=len(clientServerConns)
		if clientServerConnsLen>0 {
			num := rand.Intn(clientServerConnsLen)
			clientServerStream, err := clientServerConns[num].Open()
			//去除不存活的连接
			if err != nil {
				switch num{
				case 0:
					clientServerConns=clientServerConns[1:]
				case clientServerConnsLen-1:
					clientServerConns=clientServerConns[:clientServerConnsLen-1]
				default:
					clientServerConns=append(clientServerConns[:num-1],clientServerConns[num+1:]...)
				}
				continue
			}
			return clientServerStream
		}else{
			log.Print("there is no client server...")
			time.Sleep(5 * time.Second)
			continue
		}

	}
}


func main(){
	mode := flag.String("m", "", "client or server")
	remoteAddress := flag.String("r", "", "remote server address 1.1.1.1:1234")
	port := flag.String("p", "", "server's socks5 port")
	clientConnectPort:=flag.String("ccp", "", "client connected port")
	username:=flag.String("u", "", "socks5 username")
	password:=flag.String("pwd", "", "socks5 password")
	key:=flag.String("key", "", "connect key")

	flag.Usage = func() {
		fmt.Println("yipmanproxy -m server -p port1 -ccp port2 -key xxxx(max len 32)")
		fmt.Println("yipmanproxy -m client -r x.x.x.x:port2 -u foo -pwd bar -key xxxx(max len 32)")
	}
	flag.Parse()

	if *mode=="server"{
		if *port!="" && *clientConnectPort!=""{
			copy(encKey[:],*key)
			go listenClientServerConn(*clientConnectPort)
			listenClientClientConn(*port)
		}else{
			flag.Usage()
		}
	}else if *mode=="client"{
		if *remoteAddress!="" && *username!="" && *password !=""{
			copy(encKey[:],*key)
			clientServerServeSocks(*remoteAddress,*username,*password)
		}else{
			flag.Usage()
		}
	}else {
		flag.Usage()
	}
}
