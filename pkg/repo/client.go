/*
 * MIT License
 *
 * Copyright (c) since 2021,  flomesh.io Authors.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package repo

import (
	"fmt"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"github.com/go-resty/resty/v2"
	"k8s.io/klog/v2"
	"net/http"
	"strings"
	"sync"
	"time"
)

type PipyRepoClient struct {
	baseUrl          string
	defaultTransport *http.Transport
	httpClient       *resty.Client
	mu               sync.Mutex
}

func NewRepoClient(repoRootUrl string) *PipyRepoClient {
	return NewRepoClientWithTransport(
		repoRootUrl,
		&http.Transport{
			DisableKeepAlives:  false,
			MaxIdleConns:       10,
			IdleConnTimeout:    60 * time.Second,
			DisableCompression: false,
		})
}

func NewRepoClientWithTransport(repoRootUrl string, transport *http.Transport) *PipyRepoClient {
	return newRepoClientWithRepoRootUrlAndTransport(
		repoRootUrl,
		transport,
	)
}

func newRepoClientWithRepoRootUrlAndTransport(repoRootUrl string, transport *http.Transport) *PipyRepoClient {
	repo := &PipyRepoClient{
		baseUrl:          repoRootUrl,
		defaultTransport: transport,
	}

	repo.httpClient = resty.New().
		SetTransport(repo.defaultTransport).
		SetScheme(commons.DefaultHttpSchema).
		SetAllowGetMethodPayload(true).
		SetBaseURL(repo.baseUrl).
		SetTimeout(5 * time.Second).
		SetDebug(true).
		EnableTrace()

	return repo
}

func (p *PipyRepoClient) isCodebaseExists(path string) (bool, *Codebase) {
	resp, err := p.httpClient.R().
		SetResult(&Codebase{}).
		Get(fullRepoApiPath(path))

	if err == nil {
		switch resp.StatusCode() {
		case http.StatusNotFound:
			return false, nil
		case http.StatusOK:
			return true, resp.Result().(*Codebase)
		}
	}

	klog.Errorf("error happened while getting path %q, %#v", path, err)
	return false, nil
}

func (p *PipyRepoClient) get(path string) (*Codebase, error) {
	resp, err := p.httpClient.R().
		SetResult(&Codebase{}).
		Get(fullRepoApiPath(path))

	if err != nil {
		klog.Errorf("Failed to get path %q, error: %s", path, err.Error())
		return nil, err
	}

	return resp.Result().(*Codebase), nil
}

func (p *PipyRepoClient) createCodebase(path string) (*Codebase, error) {
	resp, err := p.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(Codebase{Version: 1}).
		Post(fullRepoApiPath(path))

	if err != nil {
		klog.Errorf("failed to create codebase %q, error: %s", path, err.Error())
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to create codebase %q, reason: %s", path, resp.Status())
	}

	codebase, err := p.get(path)
	if err != nil {
		return nil, err
	}

	return codebase, nil
}

func (p *PipyRepoClient) deriveCodebase(path, base string) (*Codebase, error) {
	exists, _ := p.isCodebaseExists(base)
	if !exists {
		return nil, fmt.Errorf("parent %q of codebase %q doesn't exists", base, path)
	}

	resp, err := p.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(Codebase{Version: 1, Base: base}).
		Post(fullRepoApiPath(path))

	if err != nil {
		klog.Errorf("Failed to derive codebase codebase: path: %q, base: %q, error: %s", path, base, err.Error())
		return nil, err
	}

	switch resp.StatusCode() {
	case http.StatusOK, http.StatusCreated:
		klog.V(5).Infof("Status code is %d, stands for success.", resp.StatusCode())
	default:
		klog.Errorf("Response contains error: %#v", resp.Status())
		return nil, fmt.Errorf("failed to derive codebase codebase: path: %q, base: %q, reason: %s", path, base, resp.Status())
	}

	klog.V(5).Infof("Getting info of codebase %q", path)
	codebase, err := p.get(path)
	if err != nil {
		klog.V(5).Infof("Failed to get info of codebase %q", path)
		return nil, err
	}

	klog.V(5).Infof("Successfully derived codebase: %#v", codebase)
	return codebase, nil
}

func (p *PipyRepoClient) GetFile(path string) (string, error) {
	resp, err := p.httpClient.R().
		Get(fullFileApiPath(path))

	if err != nil {
		klog.Errorf("Failed to get path %q, error: %s", path, err.Error())
		return "", err
	}

	result := string(resp.Body())
	klog.V(5).Infof("Content of %q:\n\n\n%s\n\n\n", path, result)

	return result, nil
}

func (p *PipyRepoClient) upsertFile(path string, content interface{}) error {
	// FIXME: temp solution, refine it later
	contentType := "text/plain"
	if strings.HasSuffix(path, ".json") {
		contentType = "application/json"
	}

	resp, err := p.httpClient.R().
		SetHeader("Content-Type", contentType).
		SetBody(content).
		Post(fullFileApiPath(path))

	if err != nil {
		klog.Errorf("error happened while trying to upsert %q to repo, %s", path, err.Error())
		return err
	}

	if resp.IsSuccess() {
		return nil
	}

	errstr := "repo server responsed with error HTTP code: %d, error: %s"
	klog.Errorf(errstr, resp.StatusCode(), resp.Status())
	return fmt.Errorf(errstr, resp.StatusCode(), resp.Status())
}

func (p *PipyRepoClient) delete(path string) {
	// DELETE, as pipy repo doesn't support deletion yet, this's not implemented
	panic("implement me")
}

// Commit the codebase, version is the current vesion of the codebase, it will be increased by 1 when committing
func (p *PipyRepoClient) commit(path string, version int64) error {
	resp, err := p.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(Codebase{Version: version + 1}).
		SetResult(&Codebase{}).
		Patch(fullRepoApiPath(path))

	if err != nil {
		return err
	}

	if resp.IsSuccess() {
		return nil
	}

	err = fmt.Errorf("failed to commit codebase %q, reason: %s", path, resp.Status())
	klog.Error(err)

	return err
}

// TODO: handle concurrent updating

func (p *PipyRepoClient) Batch(batches []Batch) error {
	if len(batches) == 0 {
		return nil
	}

	for _, batch := range batches {
		// 1. batch.Basepath, if not exists, create it
		klog.V(5).Infof("batch.Basepath = %q", batch.Basepath)
		var version = int64(-1)
		exists, codebase := p.isCodebaseExists(batch.Basepath)
		if exists {
			// just get the version of codebase
			version = codebase.Version
		} else {
			klog.V(5).Infof("%q doesn't exist in repo", batch.Basepath)
			result, err := p.createCodebase(batch.Basepath)
			if err != nil {
				klog.Errorf("Not able to create the codebase %q, reason: %s", batch.Basepath, err.Error())
				return err
			}

			klog.V(5).Infof("Result = %#v", result)

			version = result.Version
		}

		// 2. upload each json to repo
		for _, item := range batch.Items {
			fullpath := fmt.Sprintf("%s%s/%s", batch.Basepath, item.Path, item.Filename)
			klog.V(5).Infof("Creating/updating config %q", fullpath)
			klog.V(5).Infof("Content: %#v", item.Content)
			err := p.upsertFile(fullpath, item.Content)
			if err != nil {
				klog.Errorf("Upsert %q error, reason: %s", fullpath, err.Error())
				return err
			}
		}

		// 3. commit the repo, so that changes can take effect
		klog.V(5).Infof("Committing batch.Basepath = %q", batch.Basepath)
		// NOT a valid version, ignore committing
		if version == -1 {
			err := fmt.Errorf("%d is not a valid version", version)
			klog.Error(err)
			return err
		}
		if err := p.commit(batch.Basepath, version); err != nil {
			klog.Errorf("Error happened while committing the codebase %q, error: %s", batch.Basepath, err.Error())
			return err
		}
	}

	return nil
}

func (p *PipyRepoClient) DeriveCodebase(path, base string) error {
	klog.V(5).Infof("Checking if exists, codebase %q", path)
	exists, _ := p.isCodebaseExists(path)

	if exists {
		klog.V(5).Infof("Codebase %q already exists, ignore deriving ...", path)
	} else {
		klog.V(5).Infof("Codebase %q doesn't exist, deriving ...", path)
		result, err := p.deriveCodebase(path, base)
		if err != nil {
			klog.Errorf("Deriving codebase %q error: %#v", path, err)
			return err
		}
		klog.V(5).Infof("Successfully derived codebase %q", path)

		klog.V(5).Infof("Committing the changes of codebase %q", path)
		if err = p.commit(path, result.Version); err != nil {
			klog.Errorf("Committing codebase %q error: %#v", path, err)
			return err
		}
		klog.V(5).Infof("Successfully committed codebase %q", path)
	}

	return nil
}

func (p *PipyRepoClient) IsRepoUp() bool {
	_, err := p.get("/")

	if err != nil {
		return false
	}

	return true
}

func fullRepoApiPath(path string) string {
	return fmt.Sprintf("%s%s", commons.DefaultPipyRepoApiPath, path)
}

func fullFileApiPath(path string) string {
	return fmt.Sprintf("%s%s", commons.DefaultPipyFileApiPath, path)
}
