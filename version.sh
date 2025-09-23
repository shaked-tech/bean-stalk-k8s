#!/bin/bash

# Version management script for Pod Metrics Dashboard
set -e

VERSION_FILE="VERSION"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to display usage
usage() {
    echo "Usage: $0 <command> [options]"
    echo ""
    echo "Commands:"
    echo "  show                    Show current version"
    echo "  patch                   Increment patch version (x.y.Z)"
    echo "  minor                   Increment minor version (x.Y.0)"
    echo "  major                   Increment major version (X.0.0)"
    echo "  set <version>           Set specific version (e.g., 0.2.1)"
    echo ""
    echo "Options:"
    echo "  --deploy, -d            Deploy after version update"
    echo "  --help, -h              Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 patch                # Increment patch: 0.1.0 -> 0.1.1"
    echo "  $0 minor --deploy       # Increment minor and deploy: 0.1.1 -> 0.2.0"
    echo "  $0 set 0.5.0 -d         # Set version to 0.5.0 and deploy"
}

# Function to validate semver format
validate_version() {
    local version=$1
    if [[ ! $version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo -e "${RED}❌ Invalid version format: $version${NC}"
        echo -e "${YELLOW}Expected format: MAJOR.MINOR.PATCH (e.g., 0.1.2)${NC}"
        exit 1
    fi
}

# Function to get current version
get_current_version() {
    if [[ ! -f $VERSION_FILE ]]; then
        echo -e "${YELLOW}⚠️  VERSION file not found, creating with version 0.1.0${NC}"
        echo "0.1.0" > $VERSION_FILE
        echo "0.1.0"
    else
        cat $VERSION_FILE | tr -d '\n'
    fi
}

# Function to split version into components
split_version() {
    local version=$1
    IFS='.' read -r major minor patch <<< "$version"
    echo "$major $minor $patch"
}

# Function to increment version
increment_version() {
    local current_version=$1
    local increment_type=$2
    
    read -r major minor patch <<< "$(split_version "$current_version")"
    
    case $increment_type in
        patch)
            patch=$((patch + 1))
            ;;
        minor)
            minor=$((minor + 1))
            patch=0
            ;;
        major)
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        *)
            echo -e "${RED}❌ Invalid increment type: $increment_type${NC}"
            exit 1
            ;;
    esac
    
    echo "$major.$minor.$patch"
}

# Function to update version file
update_version() {
    local new_version=$1
    echo "$new_version" > $VERSION_FILE
    echo -e "${GREEN}✅ Version updated to: $new_version${NC}"
}

# Function to deploy
deploy() {
    local version=$1
    echo -e "${BLUE}🚀 Deploying version $version...${NC}"
    if [[ -f "./deploy-to-kind.sh" ]]; then
        ./deploy-to-kind.sh "$version"
    else
        echo -e "${RED}❌ deploy-to-kind.sh not found${NC}"
        exit 1
    fi
}

# Main execution
main() {
    if [[ $# -eq 0 ]]; then
        usage
        exit 1
    fi
    
    local command=$1
    local deploy_flag=false
    
    # Parse arguments
    shift
    while [[ $# -gt 0 ]]; do
        case $1 in
            --deploy|-d)
                deploy_flag=true
                shift
                ;;
            --help|-h)
                usage
                exit 0
                ;;
            *)
                if [[ $command == "set" && -z ${version_arg:-} ]]; then
                    version_arg=$1
                else
                    echo -e "${RED}❌ Unknown option: $1${NC}"
                    usage
                    exit 1
                fi
                shift
                ;;
        esac
    done
    
    local current_version
    current_version=$(get_current_version)
    
    case $command in
        show)
            echo -e "${BLUE}Current version: ${GREEN}$current_version${NC}"
            ;;
        patch|minor|major)
            local new_version
            new_version=$(increment_version "$current_version" "$command")
            echo -e "${BLUE}Updating version: ${YELLOW}$current_version${NC} → ${GREEN}$new_version${NC}"
            update_version "$new_version"
            
            if [[ $deploy_flag == true ]]; then
                deploy "$new_version"
            fi
            ;;
        set)
            if [[ -z ${version_arg:-} ]]; then
                echo -e "${RED}❌ Version required for 'set' command${NC}"
                usage
                exit 1
            fi
            
            validate_version "$version_arg"
            echo -e "${BLUE}Setting version: ${YELLOW}$current_version${NC} → ${GREEN}$version_arg${NC}"
            update_version "$version_arg"
            
            if [[ $deploy_flag == true ]]; then
                deploy "$version_arg"
            fi
            ;;
        *)
            echo -e "${RED}❌ Unknown command: $command${NC}"
            usage
            exit 1
            ;;
    esac
}

main "$@"
