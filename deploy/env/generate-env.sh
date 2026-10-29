#!/bin/sh

# Generates deploy/compose/<service>/.env files from the <service>.env.template templates
# and the shared deploy/env/.env file.
#
# Written in POSIX sh so it runs the same way in Task's built-in shell (Windows, Linux, macOS)
# and in any regular shell. Environment:
#   SERVICES    comma-separated list of services to process (required)
#   ENV_SUBST   path to the envsubst binary (default: envsubst from PATH)
#   SCRIPT_DIR  directory of this script (default: derived from $0, set it when the script is sourced)

SCRIPT_DIR="${SCRIPT_DIR:-$(cd "$(dirname "$0")" && pwd)}"
TEMPLATE_DIR="$SCRIPT_DIR"
COMPOSE_DIR="$SCRIPT_DIR/../compose"

# Use the ENV_SUBST variable if passed, otherwise the system envsubst
if [ -z "$ENV_SUBST" ]; then
  if ! command -v envsubst > /dev/null 2>&1; then
    echo "❌ Error: envsubst was not found in the system and was not passed via ENV_SUBST!"
    echo "Run the script via task env:generate"
    exit 1
  fi
  ENV_SUBST=envsubst
fi

# Load the main .env file
if [ ! -f "$SCRIPT_DIR/.env" ]; then
  echo "Error: file $SCRIPT_DIR/.env not found!"
  exit 1
fi

# Export all variables from .env so envsubst can use them
set -a
. "$SCRIPT_DIR/.env"
set +a

# Processes a template and creates the .env file of a service
process_template() {
  service=$1
  template="$TEMPLATE_DIR/${service}.env.template"
  output="$COMPOSE_DIR/${service}/.env"

  echo "Processing the template of service $service..."

  if [ ! -f "$template" ]; then
    echo "⚠️ Template $template not found, skipping..."
    return 0
  fi

  # Create the directory if it does not exist yet
  mkdir -p "$(dirname "$output")"

  # Substitute the variables in the template with envsubst
  "$ENV_SUBST" < "$template" > "$output"

  echo "✅ Created file $output"
}

# The list of services comes from the SERVICES environment variable
if [ -z "$SERVICES" ]; then
  echo "⚠️ The SERVICES variable is not set. No services to process."
  exit 0
fi

echo "🔍 Processing services: $SERVICES"

# Process the templates of all given services (split by comma)
success_count=0
skip_count=0
OLD_IFS=$IFS
IFS=','
for service in $SERVICES; do
  IFS=$OLD_IFS
  process_template "$service"
  if [ -f "$TEMPLATE_DIR/${service}.env.template" ]; then
    success_count=$((success_count + 1))
  else
    skip_count=$((skip_count + 1))
  fi
done
IFS=$OLD_IFS

if [ "$success_count" -eq 0 ]; then
  echo "⚠️ No .env file was created. Check the list of services and the templates."
else
  echo "🎉 Generation finished: $success_count files created, $skip_count templates skipped"
fi
