#!/bin/bash

# With the addition of Keystone, to use an openstack cloud you should
# authenticate against keystone, which returns a **Token** and **Service
# Catalog**.  The catalog contains the endpoint for all services the
# user/tenant has access to - including nova, glance, keystone, swift.
#
# *NOTE*: Using the 2.0 *auth api* does not mean that compute api is 2.0.  We
# will use the 1.1 *compute api*
export OS_AUTH_URL=http://104.130.131.164:5000/v2.0

# With the addition of Keystone we have standardized on the term **tenant**
# as the entity that owns the resources.
export OS_TENANT_ID=fcad67a6189847c4aecfa3c81a05783b
export OS_TENANT_NAME="demo"

# In addition to the owning entity (tenant), openstack stores the entity
# performing the action as the **user**.
export OS_USERNAME="admin"
# export OS_USERID=0416216d72d3417b88deb09b77282d90
export OS_REGION_NAME="RegionOne"

# With Keystone you pass the keystone password.
export OS_PASSWORD="devstack"

export OS_IMAGE_ID=a0656b48-5e7a-400c-a515-b1ce69d829ab
export OS_FLAVOR_ID=1
export OS_FLAVOR_ID_RESIZE=2

# Rackspace environment variables.
export RS_USERNAME="dse.ashwilson"
export RS_API_KEY="a1a7e5a0eb3c44de902e0130280219d7"
export RS_REGION="DFW"

export RS_IMAGE_ID="e19a734c-c7e6-443a-830c-242209c4d65d"
export RS_FLAVOR_ID="performance1-1"
export RS_FLAVOR_ID_RESIZED="performance1-2"
