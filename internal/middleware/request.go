package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func RequestHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		headers := make(map[string]string)
		span := trace.SpanFromContext(c.Request.Context())
		for k, v := range headers {
			c.Set(k, v)
			if span != nil && span.SpanContext().IsValid() {
				span.SetAttributes(attribute.String(fmt.Sprintf("http.header.%s", k), v))
			}
		}
		c.Next()
	}
}
