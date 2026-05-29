# Local Terraform Provider Testing

Use this workflow to test the DNScale Terraform provider locally with Terraform CLI
before publishing a release. It builds the provider from the current checkout and
uses Terraform's `dev_overrides` mechanism so Terraform runs that local binary.

Do not commit API keys, generated Terraform state, or the temporary CLI config.

## Prerequisites

- Go installed and available as `go`
- Terraform installed and available as `terraform`
- A DNScale API key exported as `DNSCALE_API_KEY`

## Build the Local Provider

Run these commands from the provider repository:

```bash
cd ~/projects/dnscale/providers/terraform-provider-dnscale

export DNSCALE_API_KEY='your-dnscale-api-key'

mkdir -p /tmp/dnscale-tf-local/bin /tmp/dnscale-tf-local/work

go build -o /tmp/dnscale-tf-local/bin/terraform-provider-dnscale
```

## Configure Terraform Dev Overrides

Create a temporary Terraform CLI config that points the published provider source
address at the local provider binary:

```bash
cat > /tmp/dnscale-tf-local/terraformrc <<'EOF'
provider_installation {
  dev_overrides {
    "dnscaleou/dnscale" = "/tmp/dnscale-tf-local/bin"
  }

  direct {}
}
EOF

export TF_CLI_CONFIG_FILE=/tmp/dnscale-tf-local/terraformrc
```

Terraform may print a warning that provider development overrides are active.
That is expected for this workflow.

## Create a Smoke Test Configuration

Use a unique `.com` test zone so repeated runs do not collide with previous
state:

```bash
cd /tmp/dnscale-tf-local/work

cat > main.tf <<'EOF'
terraform {
  required_providers {
    dnscale = {
      source = "dnscaleou/dnscale"
    }
  }
}

provider "dnscale" {}

variable "test_zone_name" {
  type = string
}

resource "dnscale_zone" "test" {
  name   = var.test_zone_name
  type   = "master"
  region = "EU"
  status = "active"
}

resource "dnscale_record" "a" {
  zone_id = dnscale_zone.test.id
  name    = "${var.test_zone_name}."
  type    = "A"
  content = "192.0.2.10"
  ttl     = 300
}

data "dnscale_records" "all" {
  zone_id = dnscale_zone.test.id

  depends_on = [dnscale_record.a]
}

output "zone_id" {
  value = dnscale_zone.test.id
}

output "record_count" {
  value = length(data.dnscale_records.all.records)
}
EOF

cat > terraform.tfvars <<EOF
test_zone_name = "tf-local-$(date +%Y%m%d%H%M%S).com"
EOF
```

## Run the Local Test

Initialize Terraform, create the test zone and record, verify idempotency, then
destroy the test resources:

```bash
terraform init
terraform apply -auto-approve
terraform plan -detailed-exitcode
terraform destroy -auto-approve
```

`terraform plan -detailed-exitcode` should exit with `0` when the provider is
idempotent and no drift is detected. It exits with `2` when Terraform sees a
diff, and with `1` on an error.

## Back-to-Back Smoke Test

To test two clean create/read/plan/destroy cycles back to back, generate a new
zone name and repeat the run:

```bash
cat > terraform.tfvars <<EOF
test_zone_name = "tf-local-$(date +%Y%m%d%H%M%S).com"
EOF

terraform apply -auto-approve
terraform plan -detailed-exitcode
terraform destroy -auto-approve
```

## Cleanup

After testing, remove the temporary workspace and clear the local environment
variables:

```bash
cd ~/projects/dnscale/providers/terraform-provider-dnscale
rm -rf /tmp/dnscale-tf-local
unset TF_CLI_CONFIG_FILE DNSCALE_API_KEY
```
