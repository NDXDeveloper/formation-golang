/* ============================================================================
   Section 10.1 : Monolithe modulaire vs microservices
   Description : LA COUTURE D'EXTRACTION en action — le jour où le catalogue
                 devient un service distant, ce client RÉSEAU satisfait la
                 MÊME interface catalogGetter qu'orders consomme : même
                 contrat, transport réseau. orders.New(catalogclient.New(url))
                 remplace orders.New(catalog.New()) — et le module orders ne
                 change pas d'une ligne. « Extraire un module devient un
                 refactor mécanique, pas une réécriture. »
   Fichier source : 01-monolithe-vs-microservices.md
   ============================================================================ */

package catalogclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/exemple/monolithe/internal/catalog"
)

type Client struct {
	base string
	hc   *http.Client
}

func New(base string) *Client {
	return &Client{base: base, hc: &http.Client{}}
}

// Get satisfait l'interface catalogGetter d'orders — implicitement,
// sans « implements » : mêmes méthodes, donc même contrat (§ 3.x).
func (c *Client) Get(ctx context.Context, id string) (catalog.Product, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+"/products/"+id, nil)
	if err != nil {
		return catalog.Product{}, err
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return catalog.Product{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return catalog.Product{}, fmt.Errorf("statut inattendu : %s", resp.Status)
	}
	var p catalog.Product
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return catalog.Product{}, err
	}
	return p, nil
}
