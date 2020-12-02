build:
	docker build -t gcr.io/russellsaw/readability-server --build-arg BUILDKIT_INLINE_CACHE=1 .

deploy: build push
	gcloud beta run deploy readability-server --image gcr.io/russellsaw/readability-server:latest

push:
	docker push gcr.io/russellsaw/readability-server
