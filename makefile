compile:
	go build -tags api -ldflags="-s -w" -o myapi main.go

compiledev:
	go build -tags api -ldflags="-s -w" -o myapidev main.go
 
assets:
	yarn install
	yarn release

webpack:
	yarn build

run: compile
	./myapi

vacuum: 
	ssh root@myapi1 'journalctl --vacuum-time=2d'
	ssh root@myapi2 'journalctl --vacuum-time=2d'
	ssh root@myapi3 'journalctl --vacuum-time=2d'

watch:
# go get github.com/cespare/reflex
# go mod download
# increase the file watch limit, might required on MacOS
	ulimit -n 1000
#	yarn install
#	yarn watch &
	reflex -s -r '\.go$$' make run

upapi:
	(cd .. && docker compose up api)

updb:
	(cd .. & docker compose exec osdb1 bash)

tarball:
	tar -czvf ../myprojectapi.tgz api WOIRTUMNSDFOEWR983745 public src templates test views go.* *.go makefile Dockerfile version.txt
	scp ../myprojectapi.tgz damian@myapi1:/opt/www/MyAPI/tmp/myprojectapi.tgz
	scp ../myprojectapi.tgz damian@myapi2:/opt/www/MyAPI/tmp/myprojectapi.tgz
	scp ../myprojectapi.tgz damian@myapi3:/opt/www/MyAPI/tmp/myprojectapi.tgz


deploy: tarball
	bash deploy/api.sh myapi1
	bash deploy/api.sh myapi2
	bash deploy/api.sh myapi3

restart:
	bash deploy/restart.sh myapi1 myproject 10
	bash deploy/restart.sh myapi2 myproject 10
	bash deploy/restart.sh myapi3 myproject 10

mkdirs:
	ssh damian@myapi1 "mkdir -p /opt/www/MyAPI/config /opt/www/MyAPI/dev/releases /opt/www/MyAPI/dev/shared"
	ssh damian@myapi2 "mkdir -p /opt/www/MyAPI/config /opt/www/MyAPI/dev/releases /opt/www/MyAPI/dev/shared"
	ssh damian@myapi3 "mkdir -p /opt/www/MyAPI/config /opt/www/MyAPI/dev/releases /opt/www/MyAPI/dev/shared"


deploydev:
	bash deploy/dev.sh myapi1
	bash deploy/dev.sh myapi2
	bash deploy/dev.sh myapi3
	 
restartdev:
	bash deploy/restart.sh myapi1 myprojectdev 5
	bash deploy/restart.sh myapi2 myprojectdev 5
	bash deploy/restart.sh myapi3 myprojectdev 5

createversion:
	git rev-list master --count > version.txt


relApi: createversion deploy restart  

release: createversion deploy deploydev restart releaseImage 

restart: restartDO

version:
	cat version.txt
	ssh root@myapi1 cat /opt/www/MyAPI/api/current/version.txt
	ssh root@myapi2 cat /opt/www/MyAPI/api/current/version.txt
	ssh root@myapi3 cat /opt/www/MyAPI/api/current/version.txt

runtest:
	(cd ~/myproject && docker compose exec db1 bash /mnt/scripts/backupDB.sh > /dev/null 2>&1)
	(cd ~/myproject && docker compose exec api bash /mnt/scripts/runtests.sh) || echo "runtests failed with status $$?"
	(cd ~/myproject && docker compose exec db1 bash /mnt/scripts/replacedb.sh > /dev/null 2>&1)

downloaddb:
	scp root@myapi1:/tmp/myproject.sql ../tmp/myproject.sql

dumpDB:
	ssh root@myapi1 su postgres -c 'pg_dump myproject > /tmp/myproject.sql'

dropdb:
	(cd ~/myproject && docker compose exec db1 bash /mnt/scripts/dropdb.sh)

replacedb:
	(cd ~/myproject && docker compose exec db1 bash /mnt/scripts/replacedb.sh)

# the crontab that dumps the database only dumps the data - not the schema
replacedata:
	(cd ~/myproject && docker compose exec db1 bash /mnt/scripts/replacedata.sh)


apilogs:
	bash deploy/apilogs.sh myapi1 myproject
	bash deploy/apilogs.sh myapi2 myproject
	bash deploy/apilogs.sh myapi3 myproject

devlogs:
	bash deploy/apilogs.sh myapi1 myprojectdev
	bash deploy/apilogs.sh myapi2 myprojectdev
	bash deploy/apilogs.sh myapi3 myprojectdev
	 

dblogs:
	bash deploy/apilogs.sh myapi1 postgresql
	bash deploy/apilogs.sh myapi2 postgresql
	bash deploy/apilogs.sh myapi3 postgresql

taillog1:
	ssh root@myapi1 journalctl -f "_SYSTEMD_UNIT=myproject.service"

taillog2:
	ssh root@myapi2 journalctl -f "_SYSTEMD_UNIT=myproject.service"

taillog3:
	ssh root@myapi3 journalctl -f "_SYSTEMD_UNIT=myproject.service"

# addDocker:
# 	bash scripts/addDocker.sh

cleanImage:
	rm -f logs/*
	rm -f myproject

buildImage: cleanImage
	docker build --platform linux/amd64 -t myapi:`cat version.txt` -f DockerFile .
	docker tag myapi:`cat version.txt` DOCKER_USERNAME/myapi:latest
	docker push DOCKER_USERNAME/myapi:latest

copyDB: downloaddb replacedata 
	(cd ~/myproject && docker compose exec db1 bash /mnt/scripts/replacedev.sh)

zipImage: 
	rm -rf ./tmp/docker
	mkdir -p ./tmp/docker

tarImage: zipImage
	mkdir -p ./tmp/docker/WOIRTUMNSDFOEWR983745
	cp -r data ./tmp/docker
	cp -r views templates ./tmp/docker
	cp WOIRTUMNSDFOEWR983745/image.yml ./tmp/docker/WOIRTUMNSDFOEWR983745/docker.yml
	cp docker-compose.yml routes.log ./tmp/docker
	cp version.txt ./tmp/docker
	tar -czf ../tmp/myapiImage.tgz -C ./tmp/docker .

copyImage: tarImage
	scp ../tmp/myapiImage.tgz damian@myapi1:/opt/www/myproject/dev/shared/
	scp ../tmp/myapiImage.tgz damian@myapi2:/opt/www/myproject/dev/shared/
	scp ../tmp/myapiImage.tgz damian@myapi3:/opt/www/myproject/dev/shared/

releaseImage: buildImage copyDB copyImage 


# dump the heap from the running process - see
# https://www.freecodecamp.org/news/how-i-investigated-memory-leaks-in-go-using-pprof-on-a-large-codebase-4bec4325e192/
heapdump:
	curl http://159.65.40.148:7008/debug/pprof/heap/dsgfa984gjasgdf84a7gfaigf > logs/heap1.out
	curl http://159.203.96.89:7008/debug/pprof/heap/dsgfa984gjasgdf84a7gfaigf > logs/heap2.out
	curl http://209.97.149.57:7008/debug/pprof/heap/dsgfa984gjasgdf84a7gfaigf > logs/heap3.out
 
pprof1:
	go tool pprof logs/heap1.out

pprof2:
	go tool pprof logs/heap2.out

pprof3:
	go tool pprof logs/heap3.out
