import rpc
import go_pb2
import go_pb2_grpc

ServerPath = '127.0.0.1:5001'

if __name__ == '__main__':
    conn = rpc.insecure_channel(ServerPath)
    client = go_pb2_grpc.ServerApiStub(channel=conn)
    resp = client.GoSayHello(go_pb2.GoSayHelloRequest(
        msg="world"
    ))
    print(resp)
