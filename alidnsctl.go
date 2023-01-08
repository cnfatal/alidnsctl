package alidnsctl

import (
	"context"
	"errors"
	"net"
	"regexp"
	"strings"

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

type Record struct {
	Domain string   `json:"domain,omitempty"`
	Type   string   `json:"type,omitempty"`
	Values []string `json:"values,omitempty"`
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

// func convertRecords(list []alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord) map[string]Record {
// 	ret := map[string]Record{}
// 	for _, item := range list {
// 		domain, rr, typ, value := tea.StringValue(item.DomainName),
// 			tea.StringValue(item.RR), tea.StringValue(item.Type),
// 			tea.StringValue(item.Value)

// 		if val, ok := ret[rr]; !ok {
// 			if record, ok := val[typ]; ok {
// 				record.Values = append(record.Values, value)
// 			} else {
// 				val[typ] = Record{Domain: rr + "." + domain, Type: typ, Values: []string{value}}
// 			}
// 		} else {
// 			ret[rr] = map[string]Record{}
// 		}
// 	}

// 	return Record{
// 		Domain: tea.StringValue(in.RR) + "." + tea.StringValue(in.DomainName),
// 		Type:   tea.StringValue(in.Type),
// 	}
// }

func (c *AliDNSCtl) ListRecords(ctx context.Context, domain, typ string) ([]*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord, error) {
	rr, domain := SplitDomain(domain)
	return c.listRecords(ctx, domain, rr, typ)
}

func (c *AliDNSCtl) listRecords(ctx context.Context, domain, rr, typ string) ([]*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord, error) {
	all := []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord{}
	page, size := 1, 100
	for {
		resp, err := c.cli.DescribeDomainRecords(&alidns.DescribeDomainRecordsRequest{
			DomainName: &domain,
			PageSize:   tea.Int64(int64(size)),
			Type:       tea.String(typ),
			PageNumber: tea.Int64(int64(page)),
			RRKeyWord:  &rr,
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
	if rr != "" {
		return filterRR(all, rr), nil
	}
	return all, nil
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

func (c *AliDNSCtl) SetRecord(ctx context.Context, fullDomain string, typ string, values []string) error {
	if len(values) == 0 {
		return nil
	}
	rr, domain := SplitDomain(fullDomain)
	exists, err := c.typeRecordsMap(ctx, domain, rr, typ)
	if err != nil {
		return err
	}

	tocreate := []string{}
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := exists[value]; ok {
			delete(exists, value)
			continue
		}
		tocreate = append(tocreate, value)
	}
	// use to be removed to update
	for _, toremove := range exists {
		if len(tocreate) != 0 {
			// update exists
			toupdate := tocreate[0]
			tocreate = tocreate[1:]
			if _, err := c.UpdateRecord(ctx, *toremove.RecordId, typ, rr, toupdate); err != nil {
				return err
			}
		} else {
			// reomove addtional
			if _, err := c.DeleteRecord(ctx, *toremove.RecordId); err != nil {
				return err
			}
		}
	}
	for _, value := range tocreate {
		if _, err := c.AddRecord(ctx, domain, typ, rr, value); err != nil {
			return err
		}
	}
	return nil
}

func (c *AliDNSCtl) typeRecordsMap(ctx context.Context, domain, rr, typ string) (map[string]*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord, error) {
	records, err := c.listRecords(ctx, domain, rr, typ)
	if err != nil {
		return nil, err
	}
	recordsmap := map[string]*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord{}
	for _, record := range records {
		recordsmap[*record.Value] = record
	}
	return recordsmap, nil
}

func (c *AliDNSCtl) DeleteRecordFromValues(ctx context.Context, fullDomain string, typ string, values []string) error {
	rr, domain := SplitDomain(fullDomain)
	recordsmap, err := c.typeRecordsMap(ctx, domain, rr, typ)
	if err != nil {
		return err
	}
	for _, value := range values {
		if value == "" {
			continue
		}
		if val, ok := recordsmap[value]; ok {
			if _, err := c.DeleteRecord(ctx, *val.RecordId); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *AliDNSCtl) AddRecord(ctx context.Context, domainName string, typ string, rr string, value string) (*alidns.AddDomainRecordResponseBody, error) {
	typ = detectTypeIfEmpty(value, typ)
	resp, err := c.cli.AddDomainRecord(&alidns.AddDomainRecordRequest{
		DomainName: &domainName, Type: &typ, RR: &rr, Value: &value,
	})
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (c *AliDNSCtl) UpdateRecord(ctx context.Context, id string, typ string, rr string, value string) (*alidns.UpdateDomainRecordResponseBody, error) {
	typ = detectTypeIfEmpty(value, typ)
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

func (c *AliDNSCtl) DeleteRecord(ctx context.Context, id string) (*alidns.DeleteDomainRecordResponseBody, error) {
	resp, err := c.cli.DeleteDomainRecord(&alidns.DeleteDomainRecordRequest{RecordId: &id})
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (c *AliDNSCtl) DeleteRecordBatch(ctx context.Context, fullDomain string, typ string) (*alidns.DeleteSubDomainRecordsResponseBody, error) {
	rr, domain := SplitDomain(fullDomain)
	resp, err := c.cli.DeleteSubDomainRecords(&alidns.DeleteSubDomainRecordsRequest{
		DomainName: &domain, RR: &rr, Type: &typ,
	})
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func filterRR(list []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord, rr string) []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord {
	ret := []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord{}
	for _, item := range list {
		if tea.StringValue(item.RR) != rr {
			continue
		}
		ret = append(ret, item)
	}
	return ret
}

func SplitDomain(full string) (string, string) {
	full = strings.TrimSuffix(full, ".")
	cnt, i := 0, len(full)-1
	for {
		if i == -1 {
			return "", full
		}
		if full[i] == '.' {
			cnt++
		}
		if cnt == 2 {
			break
		}
		i--
	}
	return full[:i], full[i+1:]
}

var NameTypeRegexp = regexp.MustCompile(`^(([a-zA-Z0-9_]|[a-zA-Z0-9_][a-zA-Z0-9_\-]*[a-zA-Z0-9_])\.)*([A-Za-z0-9_]|[A-Za-z0-9_][A-Za-z0-9_\-]*[A-Za-z0-9_](\.?))$`)

func detectTypeIfEmpty(value string, typ string) string {
	if typ != "" {
		return typ
	}
	// try ip
	if ip := net.ParseIP(value); ip != nil {
		if ip.To4() != nil {
			return "A"
		}
		if ip.To16() != nil {
			return "AAAA"
		}
	}
	// try name type
	if NameTypeRegexp.MatchString(value) {
		return "CNAME"
	}
	return ""
}
