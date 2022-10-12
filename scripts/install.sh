#!/bin/bash
# Draft CLI installer
#
# ######                                 #####  #       ### 
# #     # #####    ##   ###### #####    #     # #        #  
# #     # #    #  #  #  #        #      #       #        #  
# #     # #    # #    # #####    #      #       #        #  
# #     # #####  ###### #        #      #       #        #  
# #     # #   #  #    # #        #      #     # #        #  
# ######  #    # #    # #        #       #####  ####### ###    
#                                                               
# usage: 
#    curl -fsSL https://raw.githubusercontent.com/Azure/draft/main/scripts/install.sh | bash
set -e
set -f

log() {
    local level=$1
    shift
    echo "$(date -u $now) - $level - $*"
}

# dump uname immediately
uname -ar

log INFO "Information logged for Draft CLI."

# Try to get os release vars
# https://www.gnu.org/software/bash/manual/html_node/Bash-Variables.html
# https://stackoverflow.com/questions/394230/how-to-detect-the-os-from-a-bash-script
if [ -e /etc/os-release ]; then
    . /etc/os-release
    DISTRIB_ID=$ID
else
    if [ -e /etc/lsb-release ]; then
        . /etc/lsb-release
    fi
fi

if [ -z "${DISTRIB_ID}" ]; then
    log INFO "Trying to identify using OSTYPE var $OSTYPE "
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        DISTRIB_ID="$OSTYPE"
        B2KOS="linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        DISTRIB_ID="$OSTYPE"
        B2KOS="osx"
    elif [[ "$OSTYPE" == "cygwin" ]]; then
        DISTRIB_ID="$OSTYPE"
    elif [[ "$OSTYPE" == "msys" ]]; then
       DISTRIB_ID="$OSTYPE"
    elif [[ "$OSTYPE" == "win32" ]]; then
        DISTRIB_ID="$OSTYPE"
        B2KOS="win"
    elif [[ "$OSTYPE" == "freebsd"* ]]; then
        DISTRIB_ID="$OSTYPE"
    else
        log ERROR "Unknown DISTRIB_ID or DISTRIB_RELEASE."
        exit 1
    fi
fi

if [ -z "${DISTRIB_ID}" ]; then
    log ERROR "Unknown DISTRIB_ID or DISTRIB_RELEASE."
    exit 1
fi

log INFO "Distro Information as $DISTRIB_ID"

# set distribution specific vars
PACKAGER=
SYSTEMD_PATH=/lib/systemd/system
if [ "$DISTRIB_ID" == "ubuntu" ]; then
    PACKAGER=apt
elif [ "$DISTRIB_ID" == "debian" ]; then
    PACKAGER=apt
elif [[ $DISTRIB_ID == centos* ]] || [ "$DISTRIB_ID" == "rhel" ]; then
    PACKAGER=yum
elif [[ "$DISTRIB_ID" == "darwin"* ]]; then
    PACKAGER=brew
else
    PACKAGER=zypper
    SYSTEMD_PATH=/usr/lib/systemd/system
fi
if [ "$PACKAGER" == "apt" ]; then
    export DEBIAN_FRONTEND=noninteractive
fi

# Check JQ Processor and download if not present
check_jq_processor_present(){
  set +e
  log INFO "Checking locally installed JQ Processor version"
  jqversion=$(jq --version)
  log INFO "Locally installed JQ Processor version is $jqversion"
  if [ -z "${jqversion}" ]; then
    $PACKAGER install jq
  fi
  jqversion=$(jq --version)
  set -e
  if [ -z "${jqversion}" ]; then
    echo "Your machine donot have JQ processor installed, plaese make sure JQ Processor is installed. \n please perform $PACKAGER update and retry running scripts"
    exit 1
  fi
}


# Download draft cli stable version.
download_draft_cli_stable_version(){
  FILENAME="draft-$OS-$ARCH"
  log INFO "Starting Draft CLI Download for $FILENAME"
  DRAFTCLIVERSION=$(curl -L -s https://api.github.com/repos/Azure/draft/releases/latest | jq -r '.tag_name')
  log INFO "Starting Draft CLI Version $DRAFTCLIVERSION"
  DRAFTCLIURL="https://github.com/Azure/draft/releases/download/$DRAFTCLIVERSION/$FILENAME"
  curl -o /tmp/draftcli -fLO $DRAFTCLIURL
  chmod +x /tmp/draftcli
  log INFO "Finished Draft CLI download complete."
}

file_issue_prompt() {
  echo "If you wish us to support your platform, please file an issue"
  echo "https://github.com/Azure/draft/issues/new/choose"
  exit 1
}

copy_draft_files() {
  if [[ ":$PATH:" == *":$HOME/.local/bin:"* ]]; then
      if [ ! -d "$HOME/.local/bin" ]; then
        mkdir -p "$HOME/.local/bin"
      fi
      mv /tmp/draftcli "$HOME/.local/bin/draftcli"
      echo "installing to $HOME/.local/bin"
  else
      echo "installation target directory is write protected, run as root to override"
      sudo mv /tmp/draftcli /usr/local/bin/draftcli
      echo "installing to /usr/local/bin"
  fi
}

install() {
  ARCH=$(uname -m);
  OS=
  if [[ "$OSTYPE" == "linux"* ]]; then
      OS="linux"
  elif [[ "$OSTYPE" == "darwin"* ]]; then
      OS="darwin"
  elif [[ "$OSTYPE" == "win32" ]]; then
      OS="win"
  else
      echo "Draft cli isn't supported for your platform - $OSTYPE"
      file_issue_prompt
      exit 1
  fi

  if [[ "$ARCH" != "x86_64" ]]; then
       echo "Draft cli is only available for linux x86_64 architecture"
       file_issue_prompt
       exit 1
  fi

  if [[ "$ARCH" == "x86_64" ]]; then
      ARCH="amd64"
  fi

  check_jq_processor_present
  download_draft_cli_stable_version
  copy_draft_files
  echo "Draft CLI kubernetes installed."
}

install