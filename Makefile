all:
	go build life
	go install lifeweb
	sass ./sass/life.scss public/css/life.css

run: all
	./bin/lifeweb

profile: all
	./bin/lifeweb -cpuprofile=lifeweb.prof