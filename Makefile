install: build
	cp pipeline-runtime.so ~/.ipfs/plugins/.

build: pipeline-runtime.so

pipeline-runtime.so:
	go build -trimpath -buildmode=plugin -o pipeline-runtime.so .
	chmod +x pipeline-runtime.so

clean:
	rm pipeline-runtime.so