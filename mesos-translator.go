package main

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"time"
)

type MesosStateJson struct {
	Frameworks []struct {
		Executors []struct {
	 		Id string
			Container string
	 	}
	}
}

type MesosContainerInfo struct {
	TaskId string
	Expires int64
}

type MesosTranslator struct {
	url string
	containers map[string]*MesosContainerInfo
}

func NewMesosTranslator(url string) *MesosTranslator {
	m := &MesosTranslator{
		url: "http://" + url + "/state.json",
		containers: make(map[string]*MesosContainerInfo),
	}

	return m
}

func (m *MesosTranslator) update() {

	res, err := http.Get(m.url)

	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		panic(err.Error())
	}

	resJson := MesosStateJson{}
	if err := json.Unmarshal(body, &resJson); err != nil {
		panic(err)
	}

	// Add new entries
	for _, framework := range resJson.Frameworks {
		for _, executor := range framework.Executors {
			m.containers["mesos-" + executor.Container] = &MesosContainerInfo {
				TaskId: executor.Id,
				Expires: time.Now().Unix() + 3600,
			}
		}
	}

	// Remove expired entries
	for k, info := range m.containers {
		if info.Expires <= time.Now().Unix() {
			delete(m.containers, k)
		}
	}
}

func (m *MesosTranslator) translate(containerName string) string {
	if val, ok := m.containers[containerName]; ok {
		val.Expires = time.Now().Unix() + 3600
		return val.TaskId
	}

	m.update()

	if val, ok := m.containers[containerName]; ok {
		return val.TaskId
	} else {
		m.containers[containerName] = &MesosContainerInfo {
			TaskId: containerName,
			Expires: time.Now().Unix() + 3600,
		}

		return containerName
	}
}
