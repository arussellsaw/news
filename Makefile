build:
	docker build -t gcr.io/russellsaw/news .

deploy: build push
	gcloud beta run deploy news --image gcr.io/russellsaw/news:latest

push:
	docker push gcr.io/russellsaw/news