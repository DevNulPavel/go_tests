TEST:
	go build
	./go-peerflix -player VLC "magnet:?xt=urn:btih:DBABF77C106945FD1B19E1C3E16019B57C848C69&tr=http%3A%2F%2Fbt.t-ru.org%2Fann%3Fmagnet"

PROFILE:
	go build
	./go-peerflix -cpuprofile cpu.prof -memprofile mem.prof -player VLC "magnet:?xt=urn:btih:DBABF77C106945FD1B19E1C3E16019B57C848C69&tr=http%3A%2F%2Fbt.t-ru.org%2Fann%3Fmagnet"

PROFILE_VISUALIZE_CPU:
	go tool pprof ./cpu.prof

PROFILE_VISUALIZE_MEM:
	go tool pprof ./mem.prof

CLEAN:
	rm go-peerflix
	rm *.prof
	