
Use command `cd` to your project folder. 

Mine is `/home/rayson/Documents/workspace/gocode/src/github.com/raysonxin/arithmetic_consul_demo`


## start consul

```
sudo docker-compose -f docker/docker-compose.yml up
```


## start register-service

```
./register -consul.host localhost -consul.port 8500 -service.host 192.168.192.145 -service.port 9000
```

> Note:
> you must use you local ip to replace '192.168.192.145'.

## start discover-service