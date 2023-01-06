package alidnsctl

import (
	"context"
	"errors"

	alidns "github.com/alibabacloud-go/alidns-20150109/v4/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
)

type AliDNSCtl struct {
	cli *alidns.Client
}

type Options struct {
	AccesskeyID     string
	AccessKeySecret string
}

func New(options Options) (*AliDNSCtl, error) {
	cfg := &openapi.Config{
		AccessKeyId:     &options.AccesskeyID,
		AccessKeySecret: &options.AccessKeySecret,
		Endpoint:        tea.String("alidns.aliyuncs.com"),
	}
	cli, err := alidns.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &AliDNSCtl{cli: cli}, nil
}

func (c *AliDNSCtl) ListDomains(ctx context.Context) (any, error) {
	all := []*alidns.DescribeDomainsResponseBodyDomainsDomain{}
	page, size := 1, 100
	for {
		resp, err := c.cli.DescribeDomains(&alidns.DescribeDomainsRequest{
			PageSize:   tea.Int64(int64(size)),
			PageNumber: tea.Int64(int64(page)),
		})
		if err != nil {
			return nil, err
		}
		if len(resp.Body.Domains.Domain) == 0 {
			break
		}
		all = append(all, resp.Body.Domains.Domain...)
		page++
	}
	return all, nil
}

func (c *AliDNSCtl) GetDomain(ctx context.Context, domainName string) (*alidns.DescribeDomainInfoResponseBody, error) {
	resp, err := c.cli.DescribeDomainInfo(&alidns.DescribeDomainInfoRequest{
		DomainName:           &domainName,
		NeedDetailAttributes: tea.Bool(true),
	})
	if err != nil {
		return nil, err
	}
	ret := resp.Body
	ret.RecordLineTreeJson = nil
	return ret, nil
}

func (c *AliDNSCtl) ListRecords(ctx context.Context, doaminName string, rrkwd string) ([]*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord, error) {
	all := []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord{}
	page, size := 1, 100
	for {
		resp, err := c.cli.DescribeDomainRecords(&alidns.DescribeDomainRecordsRequest{
			DomainName: &doaminName,
			PageSize:   tea.Int64(int64(size)),
			PageNumber: tea.Int64(int64(page)),
			RRKeyWord: func() *string {
				if rrkwd == "" {
					return nil
				}
				return &rrkwd
			}(),
		})
		if err != nil {
			return nil, err
		}
		if len(resp.Body.DomainRecords.Record) == 0 {
			break
		}
		all = append(all, resp.Body.DomainRecords.Record...)
		page++
	}
	return all, nil
}

func (c *AliDNSCtl) GetRecord(ctx context.Context, id string) (*alidns.DescribeDomainRecordInfoResponseBody, error) {
	resp, err := c.cli.DescribeDomainRecordInfo(&alidns.DescribeDomainRecordInfoRequest{
		RecordId: &id,
	})
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (c *AliDNSCtl) SetRecord(ctx context.Context, domainName string, typ string, rr string, value string) (*alidns.AddDomainRecordResponseBody, error) {
	// list and match
	records, err := c.ListRecords(ctx, domainName, rr)
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		if record.RR != nil && *record.RR == rr {
			// match and update
			resp, err := c.UpdateRecord(ctx, *record.RecordId, typ, rr, value)
			if err != nil {
				return nil, err
			}
			return &alidns.AddDomainRecordResponseBody{RecordId: resp.RecordId, RequestId: resp.RequestId}, nil
		}
	}
	// create
	return c.AddRecord(ctx, domainName, typ, rr, value)
}

func (c *AliDNSCtl) AddRecord(ctx context.Context, domainName string, typ string, rr string, value string) (*alidns.AddDomainRecordResponseBody, error) {
	resp, err := c.cli.AddDomainRecord(&alidns.AddDomainRecordRequest{
		DomainName: &domainName, Type: &typ, RR: &rr, Value: &value,
	})
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (c *AliDNSCtl) UpdateRecord(ctx context.Context, id string, typ string, rr string, value string) (*alidns.UpdateDomainRecordResponseBody, error) {
	resp, err := c.cli.UpdateDomainRecord(&alidns.UpdateDomainRecordRequest{
		RecordId: &id, RR: &rr, Type: &typ, Value: &value,
	})
	if err != nil {
		sdkerr := &tea.SDKError{}
		// uptodate
		if errors.As(err, &sdkerr) && *sdkerr.Code == "DomainRecordDuplicate" {
			return &alidns.UpdateDomainRecordResponseBody{RecordId: &id}, nil
		}
		return nil, err
	}
	return resp.Body, nil
}

func (c *AliDNSCtl) SetRecordStatus(ctx context.Context, id string, disabled bool) (*alidns.SetDomainRecordStatusResponseBody, error) {
	resp, err := c.cli.SetDomainRecordStatus(&alidns.SetDomainRecordStatusRequest{
		RecordId: &id,
		Status: func() *string {
			if disabled {
				return tea.String("Disable")
			} else {
				return tea.String("Enable")
			}
		}(),
	})
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (c *AliDNSCtl) DeleteRecord(ctx context.Context, id string) (*alidns.DeleteDomainRecordResponseBody, error) {
	resp, err := c.cli.DeleteDomainRecord(&alidns.DeleteDomainRecordRequest{
		RecordId: &id,
	})
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
