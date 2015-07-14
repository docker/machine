/*
Package clcgo is an API wrapper for the CenturyLink Cloud (CLC) API V2. While
not a complete implementation of the API, it is capable of fulfilling common
use cases around provisioning and querying servers.

It will be invaluable for you to read the API documentation to understand what
is available:

https://t3n.zendesk.com/categories/20067994-API-v2-0-Beta-

This library attempts to follow the resource and JSON attribute names exactly
in naming its structs and fields.

Usage

All API interactions require authentication. You should begin by instantiating
a Client with the NewClient function. You can then authenticate with your CLC
username and password using the GetAPICredentials function. You can read more
about the details of authentication in the documentation for the APICredentials
struct and the GetAPICredentials function.

Once authenticated, the Client provides a function for fetching resources,
GetEntity, and one for saving resource, SaveEntity. Both of those functions
have example documentation around their use. Each individual resource has
documentation about its capabilities and requirements.

Development and Extension

If you want to use clcgo with an API resource that has not yet been
implemented, you should look at the documentation for the Entity, SaveEntity,
and CreationStatusProvidingEntity interfaces. Implementing one or more of those
interfaces, depending on the resource, will allow you to write your own
resources and interact with them in the standard fashion.
*/
package clcgo
