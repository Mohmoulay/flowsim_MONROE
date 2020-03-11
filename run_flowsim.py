#!/usr/bin/python
# -*- coding: utf-8 -*-

# Author: Mohamed Moulay
# Date: October 2019
# License: GNU General Public License v3
# Developed for use by the EU H2020 MONROE project

"""
QUIC In the Wild :D
"""

import datetime
#import dateutil.relativedelta
# import nettest
import sys, getopt
import time, os
import fileinput
import subprocess
import json
import netifaces
import time
import shutil
#"burset_size":10, "iter_no":10, "protocol":["QUIC"], "port":8081, "server_ip":["xx.xx.xx.xx"]
#Configuration for MONROE /monroe/config/
CONFIG_FILE = '/monroe/config'
RESULTS_DIR = "/monroe/results/"
CURRENT_DIR = os.getcwd() + "/"

try:
	  with open(CONFIG_FILE,'r') as fd:
		  config_para = json.load(fd)
		  nodeId = str(config_para["nodeid"])
		  server_ip = (config_para['server_ip'])
		  burset_size = (config_para["burset_size"])
		  iter_no = (config_para["iter_no"])
		  port = config_para["port"]
		  protocol = config_para["protocol"]
except Exception as e:
    print "Cannot retrive CONFIG_FILE {}".format(e)
    sys.exit(1)

for server in server_ip:
    for size in burset_size:
	  for iter in iter_no:
		for ports in port:
		    for protocols in protocol:
			  start = "%.6f" % time.time()
			  if protocols == "QUIC":
				cmd=["./flowsim",
					"client",
					"-Q",
					"-I",
					server,
					"-N",
					size,
					"-n",
					iter,
					"-p",
					ports]

			  elif protocols == "HTTP3":
				    cmd=["./flowsim",
					 "client",
					 "-3",
					 "-I",
					 server,
					 "-N",
					 size,
					 "-n",
					 iter,
					 "-p",
					 ports]


		          elif protocols == "HTTPS":
				  cmd=["./flowsim",
				     "client",
				     "-S",
				     "-I",
				     server,
				     "-N",
				     size,
				     "-n",
				     iter,
				     "-p",
				     ports]


				  if protocols == "HTTP":
			           cmd=["./flowsim",
					   		"client",
				     		"-H",
				     		"-I",
				     		server,
				     		"-N",
				     		size,
							"-n",
							iter,
				     		"-p",
				     		ports]
				     		

			          elif protocols == "TCP":
					 cmd=["./flowsim",
					 "client"
					"-I",
					server,
					"-N",
					size,
					"-n",
					iter,
					"-p",
					ports]
				     

			  else:
			      print "Unknown traceroute type: {}\nIgnoring........".format(protocols)
			      continue
			  output = subprocess.check_output(cmd)
            	    end = "%.6f" % time.time()
            	    filename = "FlowsimOutput_" + str(start) + "_" + str(end) + "_" + \
                        protocols + "_" + iter + "_" + size + "_" + \
                    	nodeId + ".json"
            	with open(CURRENT_DIR + filename, "w") as outputFile:
                		outputFile.write(output)


for resultFile in [fileName for fileName in os.listdir(CURRENT_DIR) if fileName.endswith(".json")]:
    # we do not copy directly the files in order to avoid possivble corruption during the
    # automatic export
    shutil.copy2(CURRENT_DIR + resultFile, RESULTS_DIR + resultFile + ".tmp")
    shutil.move(RESULTS_DIR + resultFile + ".tmp", RESULTS_DIR + resultFile)

# we allow some time for the exporter to send the files. When this script finishes the
# experiment stops right away, so if a file has not been uploaded it is lost.
time.sleep(30)



# myCmd1= 'cp server* /monroe/results/'
# os.system(mycmd1)
