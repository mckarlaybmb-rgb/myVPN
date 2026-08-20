# x-ui Integration

The backend provisions one VLESS client per user through the x-ui panel API. User creation logs in, retrieves the configured inbound, appends a client to its JSON `settings.clients` list, updates the inbound, and stores the returned subscription URL in `xray_clients.subscription_url`. A failed x-ui operation prevents persistence. A failed database insert triggers a best-effort remote removal.

## Configuration

```text
XUI_BASE_URL=http://x-ui:2053
XUI_USERNAME=admin
XUI_PASSWORD=CHANGE_ME
XUI_INBOUND_ID=1
```

`XUI_BASE_URL` should be the panel origin without a trailing slash. The client uses the login session cookie for subsequent panel API calls. The configured inbound must already exist and use VLESS settings compatible with the generated client.

## Lifecycle

- `AddUser` logs in, gets `/panel/api/inbounds/get/{id}`, appends the VLESS UUID, email, flow, and subscription ID, then posts the complete inbound to `/panel/api/inbounds/update/{id}`.
- `RemoveUser` logs in, gets the inbound, removes the matching UUID or email, and updates the inbound.
- Enable and disable operations map to add and remove operations.
- The generated subscription URL is `{XUI_BASE_URL}/sub/{uuid-without-hyphens}`.

The client logs successful add/remove operations without logging credentials or subscription secrets.
