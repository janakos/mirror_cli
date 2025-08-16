#!/bin/bash

# Example script: Monitor all mirrors and display a dashboard
# This script provides a simple monitoring dashboard for all mirrors

set -e

# Configuration
REFRESH_INTERVAL=30  # seconds
LOG_FILE="mirror_monitor.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to log messages
log() {
  echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

# Function to display status with colors
status_color() {
  case $1 in
    "STATUS_RUNNING")
      echo -e "${GREEN}RUNNING${NC}"
      ;;
    "STATUS_PAUSED")
      echo -e "${YELLOW}PAUSED${NC}"
      ;;
    "STATUS_TERMINATED")
      echo -e "${RED}TERMINATED${NC}"
      ;;
    "STATUS_FAILED")
      echo -e "${RED}FAILED${NC}"
      ;;
    "STATUS_SETUP")
      echo -e "${BLUE}SETUP${NC}"
      ;;
    "STATUS_SNAPSHOT")
      echo -e "${BLUE}SNAPSHOT${NC}"
      ;;
    *)
      echo -e "${YELLOW}$1${NC}"
      ;;
  esac
}

# Function to display the dashboard
display_dashboard() {
  clear
  echo "╔══════════════════════════════════════════════════════════════════════════════╗"
  echo "║                              PEERDB MIRROR DASHBOARD                         ║"
  echo "║                           Updated: $(date '+%Y-%m-%d %H:%M:%S')                           ║"
  echo "╚══════════════════════════════════════════════════════════════════════════════╝"
  echo ""

  # Get mirror list
  echo "📊 Mirror Overview:"
  echo "==================="
  
  if ! mirror_cli mirror list > /tmp/mirrors.txt 2>&1; then
    echo -e "${RED}❌ Failed to connect to PeerDB${NC}"
    echo "Check your configuration: mirror_cli config show"
    echo ""
    cat /tmp/mirrors.txt
    return 1
  fi

  # Count mirrors by status
  local total_mirrors=0
  local running_mirrors=0
  local paused_mirrors=0
  local failed_mirrors=0

  # Display mirror list
  echo ""
  printf "%-20s %-15s %-15s %-10s %-12s %-10s\n" "NAME" "SOURCE" "DESTINATION" "TYPE" "CREATED" "STATUS"
  echo "$(printf '%.0s-' {1..90})"

  # Parse mirror list and get status for each
  while IFS= read -r line; do
    if [[ $line =~ ^[^-]+[[:space:]]+[^[:space:]]+[[:space:]]+[^[:space:]]+[[:space:]]+[^[:space:]]+[[:space:]]+[^[:space:]]+$ ]]; then
      # Extract mirror name (first column)
      mirror_name=$(echo "$line" | awk '{print $1}')
      
      # Skip header line
      if [[ "$mirror_name" != "NAME" ]]; then
        total_mirrors=$((total_mirrors + 1))
        
        # Get detailed status
        if mirror_status=$(mirror_cli mirror status "$mirror_name" 2>/dev/null); then
          status=$(echo "$mirror_status" | grep "Status:" | awk '{print $2}')
          
          case $status in
            "STATUS_RUNNING")
              running_mirrors=$((running_mirrors + 1))
              ;;
            "STATUS_PAUSED")
              paused_mirrors=$((paused_mirrors + 1))
              ;;
            "STATUS_FAILED"|"STATUS_TERMINATED")
              failed_mirrors=$((failed_mirrors + 1))
              ;;
          esac
          
          # Display the line with status
          printf "%-20s " "$mirror_name"
          echo "$line" | awk '{printf "%-15s %-15s %-10s %-12s ", $2, $3, $4, $5}'
          status_color "$status"
          echo ""
        else
          echo "$line UNKNOWN"
        fi
      fi
    fi
  done < /tmp/mirrors.txt

  echo ""
  echo "📈 Summary:"
  echo "==========="
  echo -e "Total Mirrors:   ${BLUE}$total_mirrors${NC}"
  echo -e "Running:         ${GREEN}$running_mirrors${NC}"
  echo -e "Paused:          ${YELLOW}$paused_mirrors${NC}"
  echo -e "Failed/Stopped:  ${RED}$failed_mirrors${NC}"
  echo ""

  # Show recent activity
  if [[ -f "$LOG_FILE" ]]; then
    echo "📝 Recent Activity (last 5 entries):"
    echo "===================================="
    tail -n 5 "$LOG_FILE"
    echo ""
  fi

  echo "🔄 Auto-refresh in ${REFRESH_INTERVAL}s | Press Ctrl+C to exit"
  echo "💡 Commands: 'mirror_cli mirror status <name>' for details"
}

# Function to check for alerts
check_alerts() {
  local alerts_found=0
  
  # Check each mirror for issues
  while IFS= read -r line; do
    if [[ $line =~ ^[^-]+[[:space:]]+[^[:space:]]+[[:space:]]+[^[:space:]]+[[:space:]]+[^[:space:]]+[[:space:]]+[^[:space:]]+$ ]]; then
      mirror_name=$(echo "$line" | awk '{print $1}')
      
      if [[ "$mirror_name" != "NAME" ]]; then
        if mirror_status=$(mirror_cli mirror status "$mirror_name" 2>/dev/null); then
          status=$(echo "$mirror_status" | grep "Status:" | awk '{print $2}')
          
          case $status in
            "STATUS_FAILED")
              log "🚨 ALERT: Mirror '$mirror_name' has FAILED"
              alerts_found=1
              ;;
            "STATUS_TERMINATED")
              log "⚠️  WARNING: Mirror '$mirror_name' is TERMINATED"
              alerts_found=1
              ;;
          esac
        fi
      fi
    fi
  done < /tmp/mirrors.txt

  return $alerts_found
}

# Main monitoring loop
main() {
  log "🚀 Starting PeerDB mirror monitoring dashboard"
  
  # Check if mirror_cli is available
  if ! command -v mirror_cli &> /dev/null; then
    echo -e "${RED}❌ mirror_cli command not found${NC}"
    echo "Please ensure mirror_cli is installed and in your PATH"
    exit 1
  fi

  # Test connection
  echo "🔗 Testing PeerDB connection..."
  if ! mirror_cli peer list &> /dev/null; then
    echo -e "${RED}❌ Cannot connect to PeerDB${NC}"
    echo "Please check your configuration:"
    mirror_cli config show
    exit 1
  fi

  echo -e "${GREEN}✅ Connected to PeerDB successfully${NC}"
  sleep 2

  # Main monitoring loop
  while true; do
    display_dashboard
    
    # Check for alerts
    if check_alerts; then
      # If alerts found, you could add notification logic here
      # For example: send email, Slack message, etc.
      :
    fi
    
    # Wait for next refresh
    sleep "$REFRESH_INTERVAL"
  done
}

# Handle Ctrl+C gracefully
trap 'echo -e "\n\n${YELLOW}👋 Monitoring stopped${NC}"; log "🛑 Monitoring dashboard stopped"; exit 0' INT

# Show help if requested
if [[ "$1" == "--help" || "$1" == "-h" ]]; then
  echo "PeerDB Mirror Monitoring Dashboard"
  echo ""
  echo "Usage: $0 [options]"
  echo ""
  echo "Options:"
  echo "  -h, --help     Show this help message"
  echo "  --interval N   Set refresh interval in seconds (default: 30)"
  echo ""
  echo "Environment Variables:"
  echo "  REFRESH_INTERVAL   Refresh interval in seconds"
  echo "  LOG_FILE          Log file path (default: mirror_monitor.log)"
  echo ""
  echo "Examples:"
  echo "  $0                    # Start with default settings"
  echo "  $0 --interval 10      # Refresh every 10 seconds"
  echo "  REFRESH_INTERVAL=60 $0  # Refresh every minute"
  exit 0
fi

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --interval)
      REFRESH_INTERVAL="$2"
      shift 2
      ;;
    *)
      echo "Unknown option: $1"
      echo "Use --help for usage information"
      exit 1
      ;;
  esac
done

# Start the dashboard
main
