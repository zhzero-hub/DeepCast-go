from concurrent import futures
import time
import grpc
import train_pb2
import train_pb2_grpc

ServerPath = '127.0.0.1:5002'


# 实现 proto 文件中定 义的 GreeterServicer
class TrainApi(train_pb2_grpc.TrainApi):

    def SayHello(self, request,
                 target,
                 options=(),
                 channel_credentials=None,
                 call_credentials=None,
                 insecure=False,
                 compression=None,
                 wait_for_ready=None,
                 timeout=None,
                 metadata=None):
        return train_pb2.HelloResponse(hello="hello {msg}".format(msg=request.hello))

    def TrainStep(self, request,
                  target,
                  options=(),
                  channel_credentials=None,
                  call_credentials=None,
                  insecure=False,
                  compression=None,
                  wait_for_ready=None,
                  timeout=None,
                  metadata=None):
        return train_pb2.StateResponse()


def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    train_pb2_grpc.add_TrainApiServicer_to_server(TrainApi(), server)
    server.add_insecure_port(ServerPath)
    server.start()
    try:
        while True:
            time.sleep(60 * 60 * 24)  # one day in seconds
    except KeyboardInterrupt:
        server.stop(0)


if __name__ == "__main__":
    serve()
