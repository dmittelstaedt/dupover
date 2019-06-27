# dupover

## Build instructions
Build is based on Ansible. Ansible, Docker and Python docker is required.

        apl -i inventory -e "proxy=http://proxy:80" build.yml

Build for s390x

        apl -i inventory -e "proxy=http://proxy:80 s390x=true" build.yml
