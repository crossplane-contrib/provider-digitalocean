# Tilt >= v0.17.8 is required to handle escaping of colons in selector names and proper
# teardown of resources
load('ext://min_tilt_version', 'min_tilt_version')
min_tilt_version('0.17.8')

# We require at minimum CRD support, so need at least Kubernetes v1.16
load('ext://min_k8s_version', 'min_k8s_version')
min_k8s_version('1.16')

# Load the provider helpers
local_file = 'cluster/local/provider.Tiltfile'
if os.path.exists(local_file) == False:
    # TODO(khos2ow): change this URL when corresponding
    # PR in crossplane/crossplane gets merged.
    #
    # https://github.com/crossplane/crossplane/pull/1925
    remote_file = 'https://raw.githubusercontent.com/khos2ow/crossplane/tilt/cluster/local/provider.Tiltfile'
    local('curl -sSLo ' + local_file + ' ' + remote_file)

symbols = load_dynamic(local_file)
build_provider = symbols['build_provider']
deploy_provider = symbols['deploy_provider']

####################################################################################
# Crossplane Provider
####################################################################################
def build_deploy_providers():
    settings = {
        'args': [],
        'debug': False,
        'namespace': 'crossplane-system'
    }
    settings.update(read_json(
        "tilt-settings.json",
        default = {}
    ))

    # Make sure these value are set as follow (they are needed like this)
    settings['core_path'] = os.getcwd()
    settings['resource_deps'] = []
    settings['local_image'] = True

    name = 'provider-digitalocean'
    provider = {
        'short_name': 'digitalocean',
        'context': os.getcwd(),
        'go_main': './cmd/provider',
        'cmd_deps': [
            'go.mod',
            'go.sum',
            'apis',
            'cmd',
            'pkg'
        ],
        'crd_deps': [
            'apis'
        ],
        'crds_folder': 'package/crds',
        'package_name': 'khos2ow/provider-digitalocean:master',
        'image_name': 'khos2ow/provider-digitalocean-controller'
    }

    build_provider(name, provider, settings)
    deploy_provider(name, provider, settings)

build_deploy_providers()

####################################################################################
# Custom Tiltfiles
#
# Users may define their own Tilt customizations in tilt.d.
# This directory is excluded from git and these files will
# not be checked in to version control.
####################################################################################
def include_custom_files():
    for f in listdir("tilt.d"):
        include(f)

include_custom_files()
