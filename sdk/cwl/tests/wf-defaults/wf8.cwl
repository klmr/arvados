# Copyright (C) The Arvados Authors. All rights reserved.
#
# SPDX-License-Identifier: Apache-2.0

cwlVersion: v1.1
class: Workflow
inputs: []
outputs: []
$namespaces:
  arv: "http://arvados.org/cwl#"
requirements:
  SubworkflowFeatureRequirement: {}
steps:
  step1:
    in: []
    out: []
    run: default-dir8.cwl
