/*
 * Copyright (c) 2023, NVIDIA CORPORATION.  All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package templates

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/NVIDIA/holodeck/api/holodeck/v1alpha1"
)

const criOTemplate = `
: ${CRIO_VERSION:={{.Version}}

# Add Cri-o repo
echo "deb https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/$OS/ /" > /etc/apt/sources.list.d/devel:kubic:libcontainers:stable.list
echo "deb http://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable:/cri-o:/$RUNTIME_VERSION/$OS/ /" > /etc/apt/sources.list.d/devel:kubic:libcontainers:stable:cri-o:$VERSION.list
curl -L https://download.opensuse.org/repositories/devel:kubic:libcontainers:stable:cri-o:$RUNTIME_VERSION/$OS/Release.key | apt-key add -
curl -L https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/$OS/Release.key | apt-key add -

# Install CRI-O
apt update
apt install cri-o={{.CRIOVersion}} cri-o-runc

# Start and enable Service
systemctl daemon-reload
systemctl restart crio
systemctl enable crio
`

type CriO struct {
	Version string
}

func NewCriO(env v1alpha1.Environment) *CriO {
	return &CriO{
		Version: env.Spec.ContainerRuntime.Version,
	}
}

func (t *CriO) Execute(tpl *bytes.Buffer, env v1alpha1.Environment) error {
	criOTemplate := template.Must(template.New("crio").Parse(criOTemplate))
	if err := criOTemplate.Execute(tpl, t); err != nil {
		return fmt.Errorf("failed to execute crio template: %v", err)
	}

	return nil
}
