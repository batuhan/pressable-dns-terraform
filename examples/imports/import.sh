#!/usr/bin/env sh
set -eu

terraform import pressable_site.main 12345
terraform import pressable_site_domain.main 12345/67890
terraform import pressable_dns_record.www 111/222
terraform import pressable_site_setting.maintenance 12345/maintenance-mode

