import nmap
import argparse
import json

parser = argparse.ArgumentParser()
parser.add_argument("target", help="Domain or IP of the target to scan")
parser.add_argument("-t", "--type", help="Scan type", choices=["quick", "full"], default="quick")

args = parser.parse_args()

if args.type == "quick":
    extra_args = "-F"
elif args.type == "full":
    extra_args = ""
else:
    extra_args = ""

if args.type == "quick":
    nm = nmap.PortScanner()
else:
    nm = []
    for i in range(8):
        nm.append(nmap.PortScannerAsync())

if args.type == "quick":
    nm.scan(args.target, arguments=extra_args)

    hosts = nm.all_hosts()
    active_hosts = [host for host in hosts if nm[host].state() == "up"]
    data = {}
    for host in active_hosts:
        data[host] = {}
        for protocol in nm[host].all_protocols():
            data[host][protocol] = {}
            for port in nm[host][protocol].keys():
                data[host][protocol][port] = {
                    "name": nm[host][protocol][port]["name"]
                }
    print(json.dumps(data))
