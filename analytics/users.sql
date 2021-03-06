SELECT
  user_id,
	insertion_timestamp,
	JSON_EXTRACT_SCALAR(payload, "$.referer")
FROM
  `russellsaw.news.analytics_raw`
WHERE JSON_EXTRACT_SCALAR(payload, "$.path") = "/" AND
JSON_EXTRACT_SCALAR(payload, "$.user_agent") NOT LIKE "%bot%"
ORDER BY
 insertion_timestamp DESC
