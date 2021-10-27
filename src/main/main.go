package main

import (
	"context"
	"gee-rpc/src/registry"
	server "gee-rpc/src/server"
	xclient "gee-rpc/src/xclient"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type Foo int

type Args struct{ Num1, Num2 int }

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func startServer(registryAddr string, wg *sync.WaitGroup) {
	var foo Foo
	l, _ := net.Listen("tcp", ":0")
	server := server.NewServer()
	_ = server.Register(&foo)
	registry.HeartBeat(registryAddr, "tcp@"+l.Addr().String(), 0)
	wg.Done()
	server.Accept(l)

	// addrCh <- l.Addr().String()
	// server.Accept(l)
	// _ = server.Register(&foo)
	// server.HandleHTTP()
	// addrCh <- l.Addr().String()
	// _ = http.Serve(l, nil)
	// if err := geerpc.Register(&foo); err != nil {
	// 	log.Fatal("register error:", err)
	// }
	// //pick a free port
	// l, err := net.Listen("tcp", ":0")
	// if err != nil {
	// 	log.Fatal("network error:", err)
	// }
	// log.Println("start rpc server on", l.Addr())
	// addr <- l.Addr().String()
	// geerpc.Accept(l)
}

// func startServer(addr chan string) {
// 	// pick a free port
// 	l, err := net.Listen("tcp", ":0")
// 	if err != nil {
// 		log.Fatal("network errors: ", err)
// 	}
// 	log.Println("start rpc server on", l.Addr())
// 	addr <- l.Addr().String()
// 	geerpc.Accept(l)
// }

func (f Foo) Sleep(args Args, reply *int) error {
	time.Sleep(time.Second * time.Duration(args.Num1))
	*reply = args.Num1 + args.Num2
	return nil
}

func call(registry string) {
	d := xclient.NewGeeRegistryDiscovery(registry, 0)
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() { _ = xc.Close() }()

	// send request & receive response
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(xc, context.Background(), "call", "Foo.Sum", &Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()

	// d := xclient.NewMultiServerDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	// xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	// defer func() { _ = xc.Close() }()

	// // send request & receive response
	// var wg sync.WaitGroup
	// for i := 0; i < 5; i++ {
	// 	wg.Add(1)
	// 	go func(i int) {
	// 		defer wg.Done()
	// 		foo(xc, context.Background(), "call", "Foo.Sum", &Args{Num1: i, Num2: i * i})
	// 	}(i)
	// }
	// wg.Wait()
	// client, _ := client.DialHTTP("tcp", <-addrCh)
	// defer func() { _ = client.Close() }()

	// time.Sleep(time.Second)
	// // send request & receive response
	// var wg sync.WaitGroup
	// for i := 0; i < 5; i++ {
	// 	wg.Add(1)
	// 	go func(i int) {
	// 		defer wg.Done()
	// 		args := &Args{Num1: i, Num2: i * i}
	// 		var reply int
	// 		if err := client.Call(context.Background(), "Foo.Sum", args, &reply); err != nil {
	// 			log.Fatal("call Foo.Sum error: ", err)
	// 		}
	// 		log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
	// 	}(i)
	// }
	// wg.Wait()
}

func foo(xc *xclient.XClient, ctx context.Context, typ, serviceMethod string, args *Args) {
	var reply int
	var err error
	switch typ {
	case "call":
		err = xc.Call(ctx, serviceMethod, args, &reply)
	case "broadcast":
		err = xc.Broadcast(ctx, serviceMethod, args, &reply)
	}
	if err != nil {
		log.Printf("%s %s error: %v", typ, serviceMethod, err)
	} else {
		log.Printf("%s %s success: %d + %d = %d", typ, serviceMethod, args.Num1, args.Num2, reply)
	}
}

func broadcast(registry string) {
	d := xclient.NewGeeRegistryDiscovery(registry, 0)
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() { _ = xc.Close() }()

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(xc, context.Background(), "broadcast", "Foo.Sum", &Args{Num1: i, Num2: i * i})
			// except 2 - 5 timeout
			ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
			foo(xc, ctx, "broadcast", "Foo.Sum", &Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()

	// d := xclient.NewMultiServerDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	// xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	// defer func() { _ = xc.Close() }()

	// var wg sync.WaitGroup
	// for i := 0; i < 5; i++ {
	// 	wg.Add(1)
	// 	go func(i int) {
	// 		defer wg.Done()
	// 		foo(xc, context.Background(), "broadcast", "Foo.Sum", &Args{Num1: i, Num2: i * i})
	// 		// expect 2 - 5 timeout
	// 		ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
	// 		foo(xc, ctx, "broadcast", "Foo.Sleep", &Args{Num1: i, Num2: i * i})
	// 	}(i)
	// }
	// wg.Wait()
}

func startRegistry(wg *sync.WaitGroup) {
	l, _ := net.Listen("tcp", ":9999")
	registry.HandleHTTP()
	wg.Done()
	_ = http.Serve(l, nil)
}

func main() {
	log.SetFlags(0)
	registryAddr := "http://localhost:9999/_geerpc_/registry"
	var wg sync.WaitGroup
	wg.Add(1)
	go startRegistry(&wg)
	wg.Wait()

	time.Sleep(time.Second)
	wg.Add(2)
	go startServer(registryAddr, &wg)
	go startServer(registryAddr, &wg)
	wg.Wait()

	time.Sleep(time.Second)
	call(registryAddr)
	broadcast(registryAddr)

	// ch1 := make(chan string)
	// ch2 := make(chan string)
	// // start two servers
	// go startServer(ch1)
	// go startServer(ch2)

	// addr1 := <-ch1
	// addr2 := <-ch2

	// time.Sleep(time.Second)
	// call(addr1, addr2)
	// broadcast(addr1, addr2)
	// log.SetFlags(0)
	// ch := make(chan string)
	// go call(ch)
	// startServer(ch)
	// log.SetFlags(0)
	// addr := make(chan string)
	// go startServer(addr)
	// client, _ := client.Dial("tcp", <-addr)
	// defer func() { _ = client.Close() }()

	// time.Sleep(time.Second)
	// // send request & receive response
	// var wg sync.WaitGroup
	// for i := 0; i < 5; i++ {
	// 	wg.Add(1)
	// 	go func(i int) {
	// 		defer wg.Done()
	// 		args := fmt.Sprintf("gee rpc req %d", i)
	// 		var reply string
	// 		if err := client.Call(context.TODO(), "Foo.Sum", args, &reply); err != nil {
	// 			log.Fatal("call Foo.Sum error: ", err)
	// 		}
	// 		log.Println("reply: ", reply)
	// 	}(i)
	// }
	// wg.Wait()
}

// func startServer(addr chan string) {
// 	// pick a free port
// 	l, err := net.Listen("tcp", ":0")
// 	if err != nil {
// 		log.Fatal("network error: ", err)
// 	}
// 	log.Println("start rpc server on", l.Addr())
// 	addr <- l.Addr().String()
// 	geerpc.Accept(l)
// }

// func main() {
// 	addr := make(chan string)
// 	go startServer(addr)

// 	// in fact, following code is like a simple geerpc client
// 	conn, _ := net.Dial("tcp", <-addr)
// 	defer func() { _ = conn.Close() }()

// 	time.Sleep(time.Second)
// 	// send options
// 	_ = json.NewEncoder(conn).Encode(geerpc.DefaultOption)
// 	cc := codec.NewGobCodec(conn)
// 	// send request & receive response
// 	for i := 0; i < 5; i++ {
// 		h := &codec.Header{
// 			ServiceMethod: "Foo.Sum",
// 			Seq:           uint64(i),
// 		}
// 		_ = cc.Write(h, fmt.Sprintf("geerpc req %d", h.Seq))
// 		_ = cc.ReadHeader(h)
// 		var reply string
// 		_ = cc.ReadBody(&reply)
// 		log.Println("reply: ", reply)
// 	}
// }

// func main() {
// 	log.SetFlags(0)
// 	addr := make(chan string)
// 	go startServer(addr)
// 	client, _ := client.Dial("tcp", <-addr)
// 	defer func() { _ = client.Close() }()

// 	time.Sleep(time.Second)
// 	// send request & receive response
// 	var wg sync.WaitGroup
// 	for i := 0; i < 5; i++ {
// 		wg.Add(1)
// 		go func(i int) {
// 			defer wg.Done()
// 			args := &Args{Num1: i, Num2: i * i}
// 			var reply int
// 			if err := client.Call(context.TODO(), "Foo.Sum", args, &reply); err != nil {
// 				log.Fatal("call Foo.Sum error:", err)
// 			}
// 			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
// 		}(i)
// 	}
// 	wg.Wait()
// }
