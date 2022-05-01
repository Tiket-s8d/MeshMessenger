package main

import (
	"fmt"
	"net"
	"os"
	"bufio"
	"encoding/json"
	"strings"	
)


type Node struct {
	Name string
	Connection map[string]Connections
	Address Address
}
type Connections struct{
	IPv6 string
	Connect bool
}

type Address struct {
	IPv6 string
	Port string
}

type Package struct {
	To string
	FromName string
	FromIP string
	Data string
}

var LocalAddress string

func init(){
	iface, err := net.Interfaces()
	if err != nil {fmt.Println(err)}
	for _,v := range iface{
		x,_ := v.Addrs()
		if v.Name[0] == 'w'{
			LocalAddress = ((x[1]).String())[:len((x[1]).String())-3]
		}
	}
}


func main(){
	fmt.Print("Write your name: ")   
	name,err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {fmt.Println(err)}
	start_node := NewNode(name)
	go start_node.Multicast()
	start_node.Run(handleServer,handleClient)
}

func (node *Node)Multicast(){
	addr,err := net.ResolveUDPAddr("udp6","[ff02::1%wlp1s0]:7575")
	msgToConnect := fmt.Sprintf("%s!%s qwejhg789",node.Name,node.Address.IPv6)
	if err != nil{fmt.Println(err)}
	conn,err := net.DialUDP("udp6",nil,addr)
	if err != nil {fmt.Println(err)}
	msg := strings.ReplaceAll(msgToConnect,"\n","")
	conn.Write([]byte(msg))
	conn.Close()
}

func NewNode(name string)*Node{
	return &Node{
		Name:name,
		Connection: map[string]Connections{
			name:Connections{
				IPv6:LocalAddress,
				Connect:false,
			},
		},
		Address : Address{
			IPv6: LocalAddress,
			Port: "7575",
		},
	}
}

func (node *Node) Run(handleServer func(*Node),handleClient func(*Node)){
	go handleServer(node)
	handleClient(node)
}

func handleServer(node *Node){
	addr := net.UDPAddr{
		Port:7576,
		IP: net.ParseIP("[::]"),
	}
	listen,err := net.ListenUDP("udp6",&addr)
	if err != nil {fmt.Println(err)}
	for{
		handleConnect(node,listen)
	}

	listen.Close()

}

func(node *Node) handleMulticast(message string){
	message = strings.ReplaceAll(message,"qwejhg789","")
	splited := strings.Split(message,"!")
	for i := 0;i != len(splited);i+=2{
		node.Connection[splited[i]] = Connections{
			IPv6: splited[i+1][:len(splited[i+1])-1],
			Connect: false,
		}
	}
}

func(node *Node) handleConnect(node *Node, conn net.Conn){
	var (
		buffer = make([]byte,1024)
		message string
		pack Package
	)
	for {
		length,err := conn.Read(buffer)
		if err != nil{fmt.Println(err)}
		message = string(buffer[:length])
		if strings.Contains(message,"qwejhg789")==true {
			node.handleMulticast(message)
		}else{
			err := json.Unmarshal([]byte(message),&pack)
			if err != nil {fmt.Println(err)}
			msg := pack.FromName + ": "+pack.Data
			msg = strings.ReplaceAll(msg,"\n","")
			fmt.Print("\n"+msg)
			break
		}
	}
}



func handleClient(node *Node){
	all_commands := []string{"/exit","/print","/connect","/network","/test","/search","/help","/multi"}
	for{
		message := InputString()
		
		splited := strings.Split(message," ")
		switch splited[0]{
			case all_commands[0]: os.Exit(0)
			case all_commands[1]: node.Test() 
			case all_commands[2]: node.ConnectTo(splited[1])
			case all_commands[3]: node.PrintConnections()
			case all_commands[4]: node.Test()
			case all_commands[5]: node.Search() 
			case all_commands[6]: fmt.Println(all_commands)
			case all_commands[7]: node.Multicast()
			default:node.SendMessage(message)
		}
	}
}

func (node *Node) Test(){
	for k,v := range node.Connection{
		fmt.Println(v.Connect,k)
	}	
}

func (node *Node) Search(){
	for k,_ := range node.Connection{
		
	}
}

func (node *Node) PrintConnections(){
	for v,k := range node.Connection{
		x := fmt.Sprintf("%s:%t",v,k.Connect)
		x = strings.ReplaceAll(x,"\n","")
		fmt.Println(x);
	}
}
func (node *Node) ConnectTo(addr string){
	for k,_ := range node.Connection{
		if addr == k{
			if entry,ok := node.Connection[addr];ok{
				entry.Connect = true
				node.Connection[addr] = entry
			}
		}	
	}
}

func (node *Node) SendMessage(msg string,port string){
	for k,v := range node.Connection{
		if v.Connect == true{
			pack := Package{
				To:k,
				FromName:node.Name,
				FromIP:node.Address.IPv6,
				Data:msg,
			}
			msg_to_send,err := json.Marshal(pack)
			if err != nil{fmt.Println(err)}
			addr := "["+v.IPv6+"%wlp1s0]:"+port
			ip,err := net.ResolveUDPAddr("udp6",addr)
			if err != nil{fmt.Println(err)}
			conn,err := net.DialUDP("udp6",nil,ip)
			conn.Write([]byte(msg_to_send))
			conn.Close()
		}
	}
}

func InputString() string{
	fmt.Print("me: ")
	msg,err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {fmt.Println(err)}
	return strings.Replace(msg,"\n","",-1)
}
