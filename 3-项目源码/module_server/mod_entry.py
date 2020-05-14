import mod_service
import mod_manage
import mod_redirect
import pathlib
from concurrent import futures
import service_pb2_grpc
import grpc
import argparse
import sys
import logging

def main(root_path, port):
    root_path = pathlib.Path(root_path)
    module_manager = mod_manage.ModuleManage(root_path)
    redirect_manager = mod_redirect.ModuleRedirect()
    redirect_manager.register_urls(module_manager.get_redirect_urls())
    logging.basicConfig()
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    service = mod_service.NotifyService(module_manager,redirect_manager)
    service_pb2_grpc.add_NotifyServiceServicer_to_server(service, server)
    server.add_insecure_port('[::]:{}'.format(port))
    server.start()
    try:
        server.wait_for_termination()
    finally:
        module_manager.save_state()

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument('-r', '--root_path', action='store',
                        required=True, dest='root_path',
                        help='specify the root path')
    parser.add_argument('-p', '--port', action='store',
                        type=int, required=True, dest='port',
                        help='specify the port to be used')
    args = parser.parse_args()
    sys.exit(main(args.root_path, args.port))
