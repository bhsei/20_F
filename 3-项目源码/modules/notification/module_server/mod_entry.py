import mod_service
from concurrent import futures
import service_pb2_grpc
import grpc
import argparse
import sys

def main(root_path, port):
    mod_service.mod_service_init(root_path, port)
    server = grpc.server(futures.ThreadPoolExecutor(max_workers = 10))
    service_pb2_grpc.add_NotifyServiceServicer_to_server(mod_service.NotifyService(), server)
    server.add_insecure_port('[::]:{}'.format(port))
    server.start()
    server.wait_for_termination()

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument('-r', '--root_path', action = 'store',
            required = True, dest = 'root_path',
            help = 'specify the root path')
    parser.add_argument('-p', '--port', action = 'store',
            type = int, required = True, dest = 'port',
            help = 'specify the port to be used')
    args = parser.parse_args()
    sys.exit(main(args.root_path, args.port))

