#!/bin/bash

PROJ="dev"

oc delete -n $PROJ all -l app=gogsfilter

exit 0
