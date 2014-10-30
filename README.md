# Diego Cluster Simulations

A collection of simulations of Diego that run on Diego.

## Auction Scenarios

In practice, spinning up hundreds of external processes to simulate a distributed auction fails hard when one is running such a simulation on a single desktop/laptop computer.  The small-scale scenarios encoded in the [auction](http://github.com/cloudfoundry-incubator/auction) simulation work fine, but the larger-scale scenarios begin to fail because of resource constraints that are unrealistic and not endemic to the auction itself.  In short: the best to way to build confidence in the simulation (particularly a large-scale scenario) is to actually run the simulation on a cluster.

This can be done with Diego.  When Diego gets an API it will be possible to do this by simply pointing the simulation at Diego's API endpoint.  Until then, you must follow these steps:

1. Run ./build.sh in `autioneer-lite` and `rep-lite`
2. Use `github.com/pivotal-cf-experimental/veritas` to desire N LRPs for `auctioneer-lite` and N LRPs for `rep-lite` (note: you'll need to set NATSUSERNAME, NATSPASSWORD, NATSADDRESSES, ETCDCLUSTER:

```bash
cat > desired_lrp_rep-lite.json <<EOF
{  
   "process_guid":"rep-lite-1",
   "domain":"veritas",
   "root_fs":"",
   "instances":1,
   "stack":"lucid64",
   "actions":[  
      {  
         "action":"download",
         "args":{  
            "from":"http://onsi-public.s3.amazonaws.com/rep-lite.tar.gz",
            "to":".",
            "extract":true,
            "cache_key":"rep-lite"
         }
      },
      {  
         "action":"download",
         "args":{  
            "from":"PLACEHOLDER_FILESERVER_URL/v1/static/linux-circus/linux-circus.tgz",
            "to":"/tmp/circus",
            "extract":true,
            "cache_key":"linux-circus"
         }
      },
      {  
         "action":"parallel",
         "args":{  
            "actions":[  
               {  
                  "action":"run",
                  "args":{  
                     "path":"./rep-lite",
                     "args":[  
                        "-repGuid=rep-lite-1",
                        "-natsUsername=NATSUSERNAME",
                        "-natsPassword=NATSPASSWORD",
                        "-natsAddresses=NATSADDRESSES"
                     ],
                     "env":[  

                     ],
                     "timeout":0,
                     "resource_limits":{  

                     }
                  }
               },
               {  
                  "action":"monitor",
                  "args":{  
                     "action":{  
                        "action":"run",
                        "args":{  
                           "path":"/tmp/circus/spy",
                           "args":[  
                              "-addr=:8080"
                           ],
                           "env":null,
                           "timeout":0,
                           "resource_limits":{  

                           }
                        }
                     },
                     "healthy_hook":{  
                        "method":"PUT",
                        "url":"http://127.0.0.1:20515/lrp_running/rep-lite-1/PLACEHOLDER_INSTANCE_INDEX/PLACEHOLDER_INSTANCE_GUID"
                     },
                     "unhealthy_hook":{  
                        "method":"",
                        "url":""
                     },
                     "healthy_threshold":1,
                     "unhealthy_threshold":1
                  }
               }
            ]
         }
      }
   ],
   "disk_mb":256,
   "memory_mb":256,
   "ports":[  
      {  
         "container_port":8080
      }
   ],
   "routes":[  
      "rep-lite-1.diego-1.cf-app.com"
   ],
   "log":{  
      "guid":"rep-lite-1",
      "source_name":"VRT"
   }
}
EOF

cat > desired_lrp_auctioneer-lite.json <<EOF
{  
   "process_guid":"auctioneer-lite-1",
   "domain":"veritas",
   "root_fs":"",
   "instances":1,
   "stack":"lucid64",
   "actions":[  
      {  
         "action":"download",
         "args":{  
            "from":"http://onsi-public.s3.amazonaws.com/auctioneer-lite.tar.gz",
            "to":".",
            "extract":true,
            "cache_key":"auctioneer-lite"
         }
      },
      {  
         "action":"download",
         "args":{  
            "from":"PLACEHOLDER_FILESERVER_URL/v1/static/linux-circus/linux-circus.tgz",
            "to":"/tmp/circus",
            "extract":true,
            "cache_key":"linux-circus"
         }
      },
      {  
         "action":"parallel",
         "args":{  
            "actions":[  
               {  
                  "action":"run",
                  "args":{  
                     "path":"./auctioneer-lite",
                     "args":[  
                        "-timeout=1s",
                        "-etcdCluster=ETCDCLUSTER",
                        "-natsUsername=NATSUSERNAME",
                        "-natsPassword=NATSPASSWORD",
                        "-natsAddresses=NATSADDRESSES"
                     ],
                     "env":[],
                     "timeout":0,
                     "resource_limits":{  

                     }
                  }
               },
               {  
                  "action":"monitor",
                  "args":{  
                     "action":{  
                        "action":"run",
                        "args":{  
                           "path":"/tmp/circus/spy",
                           "args":[  
                              "-addr=:8080"
                           ],
                           "env":null,
                           "timeout":0,
                           "resource_limits":{  

                           }
                        }
                     },
                     "healthy_hook":{  
                        "method":"PUT",
                        "url":"http://127.0.0.1:20515/lrp_running/auctioneer-lite-1/PLACEHOLDER_INSTANCE_INDEX/PLACEHOLDER_INSTANCE_GUID"
                     },
                     "unhealthy_hook":{  
                        "method":"",
                        "url":""
                     },
                     "healthy_threshold":1,
                     "unhealthy_threshold":1
                  }
               }
            ]
         }
      }
   ],
   "disk_mb":256,
   "memory_mb":256,
   "ports":[  
      {  
         "container_port":8080
      }
   ],
   "routes":[  
      "auctioneer-lite-1.diego-1.cf-app.com"
   ],
   "log":{  
      "guid":"auctioneer-lite-1",
      "source_name":"VRT"
   }
}
EOF

for i in {1..400}; do sed "s/rep-lite-1/rep-lite-$i/g" desired_lrp_rep-lite.json > temp.json; veritas submit-lrp temp.json; done
for i in {1..400}; do sed "s/auctioneer-lite-1/auctioneer-lite-$i/g" desired_lrp_auctioneer-lite.json > temp.json; veritas submit-lrp temp.json; done

for i in {1..400}; do veritas remove-lrp rep-lite-$i; done
for i in {1..400}; do veritas remove-lrp auctioneer-lite-$i; done
```

3. Once this is done, you can run `ginkgo` under `auctionscenarios` to run the simulation on the cluster!
4. Compiling auctionscenarios yields a binary that runs through a number of cases.  You can push this binary, along with the test suite (`ginkgo build`) to the cluster to run a (very large, timeconsuming) simulation.