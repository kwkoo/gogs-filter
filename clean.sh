#!/bin/bash

devProject="dev"

oc delete -n $devProject all -l app=gogsfilter
