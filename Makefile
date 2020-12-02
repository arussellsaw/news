build:
	docker build -t gcr.io/russellsaw/news --build-arg BUILDKIT_INLINE_CACHE=1 .

deploy: build push
	gcloud beta run deploy news --image gcr.io/russellsaw/news:latest

push:
	docker push gcr.io/russellsaw/news
