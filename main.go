package main

import (
	"fmt"
	"strconv"
	"net"
	"os"
	"bufio"
	"encoding/json"
	"strings"	
)


type Node struct {
	Name string
	Connection map[string][]Connections
	Address Address
}
type Connections struct{
	IPv6 string
	Name string
	Connect bool
}

type Address struct {
	IPv6 string
	Port string
}

type Package struct {
	Route []string
	To string
	FromName string
	FromIP string
	Data string
}

var PORT_FOR_SEND string
var LISTEN_PORT string

func main(){
	fmt.Print("Write your name: ")   
	name,err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {fmt.Println(err)}
	
	fmt.Print("Write your port send: ")
	PORT_FOR_SEND,err = bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {fmt.Println(err)}
	PORT_FOR_SEND = strings.ReplaceAll(PORT_FOR_SEND,"\n","")
	
	fmt.Print("Write your port listen: ")
	LISTEN_PORT,err = bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {fmt.Println(err)}
	LISTEN_PORT = strings.ReplaceAll(LISTEN_PORT,"\n","")
	
	start_node := NewNode(name)
	go start_node.Multicast() 
	start_node.Run(handleServer,handleClient)
}

// Make new node with filed strings
func NewNode(name string)*Node{
	var LocalAddress string
	iface, err := net.Interfaces()
	if err != nil {fmt.Println(err)}
	for _,v := range iface{
		x,_ := v.Addrs()
		if v.Name[0] == 'w'{
			LocalAddress = ((x[1]).String())[:len((x[1]).String())-3]
		}
	}
	
	return &Node{
		Name:name,
		
		Address : Address{
			IPv6: LocalAddress,
			Port: PORT_FOR_SEND,
		},
	}
}
// Start two handle methods
func (node *Node) Run(handleServer func(*Node),handleClient func(*Node)){
	go handleServer(node)
	handleClient(node)
}

// Listen port for new message
func handleServer(node *Node){
	port,err := strconv.Atoi(LISTEN_PORT)
	if err != nil {fmt.Println(err)}
	addr := net.UDPAddr{
		Port:port,
		IP: net.ParseIP("[::]"),
	}
	listen,err := net.ListenUDP("udp6",&addr) // Address with port who we listen
	if err != nil {fmt.Println(err)}
	for{
		handleConnect(node,listen)
	}

	listen.Close()

}

func (node *Node)Multicast(){
	var pack Package

	addr,err := net.ResolveUDPAddr("udp6",("[ff02::1%wlan0]:"+node.Address.Port)) // Get address for send message
	if err != nil{fmt.Println(err)}
	pack = Package{
		To:"All",
		FromName:node.Name,
		FromIP:node.Address.IPv6,
		Data:"M0aIbHfcKeMg5rcCh3NDaflcC3xLIdWN",
	}
	msg,err := json.Marshal(pack)
	if err != nil{fmt.Println(err)}
	conn,err := net.DialUDP("udp6",nil,addr) // Make connections to addr
	if err != nil {fmt.Println(err)}

	conn.Write([]byte(msg))
	conn.Close()
}
// Processing of multicast message
func(node *Node) MulticastProcessing(message Package){
	if _,ok := node.Connection[message.FromName];ok == false{ 
		KeyConnect := "M0aIbHfcKeMg5rcCh3NDaflcC3xLIdWN"
		node.Connection[message.FromName] = append(node.Connection[message.FromName],Connections{
			IPv6: message.FromIP,
			Connect: true,
		})
		node.SendMessage(KeyConnect)
		node.Connection[message.FromName][0] = Connections{
			IPv6: message.FromIP,
			Connect: false,
		}
	}
}

// Processing of all messages
func handleConnect(node *Node, conn net.Conn){
	var (
		buffer = make([]byte,1024)
		message string
		pack Package
	)
	for {
		length,err := conn.Read(buffer)
		if err != nil{fmt.Println(err)}
		message = string(buffer[:length])

		err = json.Unmarshal([]byte(message),&pack)
		if err != nil {fmt.Println(err)}
		if strings.Contains(pack.Data,"M0aIbHfcKeMg5rcCh3NDaflcC3xLIdWN"){
			node.MulticastProcessing(pack)
		}else{
			msg := pack.FromName + ": "+pack.Data
			msg = strings.ReplaceAll(msg,"\n","")
			fmt.Print("\n"+msg)
			break
		}
	}
}



// Processing the messages what user write 
func handleClient(node *Node){
	all_commands := []string{"/exit","/print","/connect","/network","/test","/search","/help","/multi"}
	for{
		message := InputString()
		
		splited := strings.Split(message," ")
		switch splited[0]{
			case all_commands[0]: os.Exit(0)
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
	for _,k := range node.Connection{
		fmt.Println(k)
	}
}

func (node *Node) Search(){
}

// Print all connections users
func (node *Node) PrintConnections(){
//	for v,k := range node.Connection{
//	}
}

// Make connect to user who we know, and send message to him
func (node *Node) ConnectTo(addr string){
//	for k,_ := range node.Connection{
//		if (strings.ReplaceAll(k,"\n","") == addr){
//			ent := node.Connection[k]
//			ent.Connect = true
//			node.Connection[k] = ent
//		}
//	}
}

func (node *Node) SendMessage(msg string){
//	for k,v := range node.Connection{
//		if v.Connect == true{
//			pack := Package{
//				To:k,
//				FromName:node.Name,
//				FromIP:node.Address.IPv6,
//				Data:msg,
//			}
//			msg_to_send,err := json.Marshal(pack)
//			if err != nil{fmt.Println(err)}
//		}
//	}
	for n,v := range node.Connection{
		for _,k := range v{
			if k.Connect == true{
				pack := Package{
					To:n,
					FromName:node.Name,
					FromIP:node.Address.IPv6,
					Data:msg,
				}
				msg_to_send,err := json.Marshal(pack)
				if err != nil{fmt.Println(err)}

				addr := "["+k.IPv6+"%wlan0]:"+PORT_FOR_SEND
				ip,err := net.ResolveUDPAddr("udp6",addr)
				if err != nil{fmt.Println(err)}
				conn,err := net.DialUDP("udp6",nil,ip)
				conn.Write([]byte(msg_to_send))
				conn.Close()
			}
		}	
	}
}

func InputString() string{
	fmt.Print("me: ")
	msg,err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {fmt.Println(err)}
	return strings.ReplaceAll(msg,"\n","")
	
}
