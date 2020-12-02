const { Readability } = require('@mozilla/readability');
const JSDOM = require('jsdom').JSDOM;
const http = require('http');
const fetch = require('node-fetch');
const createDOMPurify = require('dompurify');
const url = require('url');

const handler = function (req, res) {
	const articleURL = new URL(req.url, "http://localhost:9090/").searchParams.get("url")
	fetch(articleURL)
		.then(res => res.text())
		.then(body => {
			var doc = new JSDOM(body, {
				url: articleURL
			});
			const DOMPurify = createDOMPurify(doc.window);
			let reader = new Readability(doc.window.document);
			let article = reader.parse();
			let clean = DOMPurify.sanitize(article.content, {SAFE_FOR_TEMPLATES: true});
			res.setHeader("Content-Type", "application/json")
			res.end(JSON.stringify({
				body: clean,
				body_text: article.textContent
			}))
		})
}

const server = http.createServer(handler);
server.listen(8080);