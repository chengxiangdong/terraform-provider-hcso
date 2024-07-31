package vpc

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tidwall/gjson"

	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils"

	"github.com/huaweicloud/terraform-provider-hcso/internal/helper/httphelper"
	"github.com/huaweicloud/terraform-provider-hcso/internal/helper/schemas"
)

func DataSourceVpcRoutes() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVpcRoutesRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `Specifies the region in which to query the resource. If omitted, the provider-level region will be used.`,
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: `Specifies the route type.`,
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: `Specifies the ID of the VPC to which the route belongs.`,
			},
			"destination": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: `Specifies the route destination.`,
			},
			"routes": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: `The list of routes.`,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The route ID.`,
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The route type.`,
						},
						"vpc_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The ID of the VPC to which the route belongs.`,
						},
						"destination": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The route destination.`,
						},
						"nexthop": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The next hop of the route.`,
						},
					},
				},
			},
		},
	}
}

type RoutesDSWrapper struct {
	*schemas.ResourceDataWrapper
	Config *config.Config
}

func newRoutesDSWrapper(d *schema.ResourceData, meta interface{}) *RoutesDSWrapper {
	return &RoutesDSWrapper{
		ResourceDataWrapper: schemas.NewSchemaWrapper(d),
		Config:              meta.(*config.Config),
	}
}

func dataSourceVpcRoutesRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	wrapper := newRoutesDSWrapper(d, meta)
	lisVpcRouRst, err := wrapper.ListVpcRoutes()
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := uuid.GenerateUUID()
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	err = wrapper.listVpcRoutesToSchema(lisVpcRouRst)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// @API VPC GET /v2.0/vpc/routes
func (w *RoutesDSWrapper) ListVpcRoutes() (*gjson.Result, error) {
	client, err := w.NewClient(w.Config, "vpc")
	if err != nil {
		return nil, err
	}

	uri := "/v2.0/vpc/routes"
	params := map[string]any{
		"type":        w.Get("type"),
		"vpc_id":      w.Get("vpc_id"),
		"destination": w.Get("destination"),
	}
	params = utils.RemoveNil(params)
	return httphelper.New(client).
		Method("GET").
		URI(uri).
		Query(params).
		MarkerPager("routes", "routes[*].id | [-1]", "marker").
		Request().
		Result()
}

func (w *RoutesDSWrapper) listVpcRoutesToSchema(body *gjson.Result) error {
	d := w.ResourceData
	mErr := multierror.Append(nil,
		d.Set("region", w.Config.GetRegion(w.ResourceData)),
		d.Set("routes", schemas.SliceToList(body.Get("routes"),
			func(route gjson.Result) any {
				return map[string]any{
					"id":          route.Get("id").Value(),
					"type":        route.Get("type").Value(),
					"vpc_id":      route.Get("vpc_id").Value(),
					"destination": route.Get("destination").Value(),
					"nexthop":     route.Get("nexthop").Value(),
				}
			},
		)),
	)
	return mErr.ErrorOrNil()
}