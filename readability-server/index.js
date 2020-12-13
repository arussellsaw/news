const { Readability } = require('@mozilla/readability');
const JSDOM = require('jsdom').JSDOM;
const http = require('http');
const fetch = require('node-fetch');
const createDOMPurify = require('dompurify');
const url = require('url');

const articleURL = process.argv[2];
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
		console.log(JSON.stringify({
			body: clean,
			body_text: article.textContent
		}))
	}).catch(error => {
		console.log(error)
		process.exit(1)
	})
