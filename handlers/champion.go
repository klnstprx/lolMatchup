package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/klnstprx/lolMatchup/components"
	"github.com/klnstprx/lolMatchup/config"
	"github.com/klnstprx/lolMatchup/models"
	"github.com/klnstprx/lolMatchup/renderer"
)

func HomeGET(c *gin.Context) {
	r := renderer.New(c.Request.Context(), http.StatusOK, components.Home())
	c.Render(http.StatusOK, r)
}

func capitlizeFirstLetter(word string) string {
	runes := []rune(word)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func ChampionGET(c *gin.Context) {
	name, ok := c.GetQuery("champion")
	if !ok || len(name) == 0 {
		c.String(http.StatusBadRequest, "Error geting champion name from query param.")
		config.App.Logger.Error("Champion name empty!", c.Request.URL)
		return
	}

	name = capitlizeFirstLetter(name)

	targetURL := fmt.Sprintf(config.App.DDragonURLData+"%s.json", name)
	config.App.Logger.Info("Querying...", "url", targetURL)
	resp, err := http.Get(targetURL)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching data for champion %s: %v", name, err)
		config.App.Logger.Errorf("Error fetching data for champion %s: %v", name, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.String(http.StatusNotFound, "Champion %s not found", name)
		config.App.Logger.Errorf("Champion %s not found", name)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error reading response body")
		config.App.Logger.Errorf("Error reading response body")
		return
	}

	var root models.Root
	err = json.Unmarshal(body, &root)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error parsing JSON data: %v", err)
		config.App.Logger.Errorf("Error parsing JSON data: %v", err)
		return
	}

	var champion models.Champion
	found := false
	for _, champ := range root.Data {
		champion = champ
		found = true
		break
	}

	if !found {
		c.String(http.StatusNotFound, "Champion %s data not found in response", name)
		config.App.Logger.Errorf("Champion %s data not found in response", name)
		return
	}

	r := renderer.New(c.Request.Context(), http.StatusOK, components.ChampionComponent(champion))
	c.Render(http.StatusOK, r)
}
