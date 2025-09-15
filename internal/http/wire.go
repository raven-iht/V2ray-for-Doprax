package httpserver

import (
	"net/http"

	echo "github.com/labstack/echo/v4"

	"example.com/decision-backend/internal/engine"
)

var defaultEngine = newDefaultEngine()

func newDefaultEngine() *engine.Engine {
	e := engine.New()
	// sample: pricing/approveOrder
	e.Register("pricing", "approveOrder", func(in engine.Input) (engine.Output, error) {
		basket, _ := in["basketTotal"].(float64)
		country, _ := in["country"].(string)
		approved := basket <= 300 && country != "SANCTIONED"
		return engine.Output{"approved": approved, "limit": 300.0}, nil
	})
	return e
}

func evaluateDecision(c echo.Context) error {
	model := c.Param("model")
	decision := c.Param("decision")

	var in map[string]any
	if err := c.Bind(&in); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
	}

	out, found, err := defaultEngine.Evaluate(model, decision, engine.Input(in))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if !found {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "model or decision not found"})
	}
	return c.JSON(http.StatusOK, out)
}
