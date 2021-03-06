package kvpaxos

import "net/rpc"
import "fmt"
import "math/big"
import "crypto/rand"
import "time"

type Clerk struct {
  servers []string
  // You will have to modify this struct.
}


func MakeClerk(servers []string) *Clerk {
  ck := new(Clerk)
  ck.servers = servers
  // You'll have to add code here.
  return ck
}

//
// call() sends an RPC to the rpcname handler on server srv
// with arguments args, waits for the reply, and leaves the
// reply in reply. the reply argument should be a pointer
// to a reply structure.
//
// the return value is true if the server responded, and false
// if call() was not able to contact the server. in particular,
// the reply's contents are only valid if call() returned true.
//
// you should assume that call() will time out and return an
// error after a while if it doesn't get a reply from the server.
//
// please use call() to send all RPCs, in client.go and server.go.
// please don't change this function.
//
func call(srv string, rpcname string,
          args interface{}, reply interface{}) bool {
  c, errx := rpc.Dial("unix", srv)
  if errx != nil {
    return false
  }
  defer c.Close()
    
  err := c.Call(rpcname, args, reply)
  if err == nil {
    return true
  }

  fmt.Println(err)
  return false
}

func nrand() int64 {
  max := big.NewInt(int64(1) << 62)
  bigx, _ := rand.Int(rand.Reader, max)
  x := bigx.Int64()
  return x
}

//
// fetch the current value for a key.
// returns "" if the key does not exist.
// keeps trying forever in the face of all other errors.
//
func (ck *Clerk) Get(key string) string {
  // You will have to modify this function.
  args := &GetArgs{}
  args.Key = key
  args.Id = nrand()
  var reply GetReply

  dargs := &DoneArgs{}
  dargs.Id = args.Id
  dreply := &DoneReply{}

  DPrintf("Client: Get [%s], request %v\n", key, args.Id)
  for done := false; !done; {
    for _, srv := range ck.servers {
      DPrintf("Client call to server %s, request %v\n", srv, args.Id)
      reply = GetReply{}
      ok := call(srv, "KVPaxos.Get", args, &reply)

      if ok {
        if reply.Err == OK || reply.Err == ErrNoKey || reply.Err == ErrDuplicate {
          call(srv, "KVPaxos.Done", dargs, dreply)
          done = true
          break
        }
      } else {
        fmt.Errorf("Get RPC failed from client to %s, request %v\n", srv, args.Id)
      }
      time.Sleep(PaxosWaitTime)
    }
  }

  DPrintf("Client finished: Get {%s: %s}, request %v\n", key, reply.Value, args.Id)
  return reply.Value
}

//
// set the value for a key.
// keeps trying until it succeeds.
//
func (ck *Clerk) PutExt(key string, value string, dohash bool) string {
  // You will have to modify this function.
  args := &PutArgs{}
  args.Key = key
  args.Value = value
  args.DoHash = dohash
  args.Id = nrand()
  var reply PutReply

  dargs := &DoneArgs{}
  dargs.Id = args.Id
  dreply := &DoneReply{}

  DPrintf("Client: Put {%s: %s}, request %v\n", key, value, args.Id)
  for done := false; !done; {
    for _, srv := range ck.servers {
      DPrintf("Client call to server %s, request %v\n", srv, args.Id)
      reply = PutReply{}
      ok := call(srv, "KVPaxos.Put", args, &reply)

      if ok {
        if reply.Err == OK || reply.Err == ErrDuplicate {
          call(srv, "KVPaxos.Done", dargs, dreply)
          done = true
          break
        }
      } else {
        fmt.Errorf("Put RPC failed from client to %s, request %v\n", srv, args.Id)
      }
      time.Sleep(PaxosWaitTime)
    }
  }

  DPrintf("Client finished: Put {%s: %s}, request %v\n", key, value, args.Id)
  return reply.PreviousValue
}

func (ck *Clerk) Put(key string, value string) {
  ck.PutExt(key, value, false)
}
func (ck *Clerk) PutHash(key string, value string) string {
  v := ck.PutExt(key, value, true)
  return v
}
