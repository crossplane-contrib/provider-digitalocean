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
#
# Users can use tilt-settings.json to pass additional configuration and
# parameters for building providers.
#
# Example content of tilt-settings.json:
#
# {
#     "args": [],                        // args to pass to pod
#     "debug": false,                    // enable debug mode
#     "namespace": "crossplane-system"   // namespace to deploy provider into
# }
####################################################################################
def build_deploy_provider():
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

    if os.path.exists('tilt-provider.json') == False:
        print('Warning: tilt-provider.json is missing!')
        return

    provider_name = 'provider-digitalocean'

    provider = {}
    provider.update(read_json(
        'tilt-provider.json',
        default = {}
    ))

    provider['short_name'] = provider_name.replace('provider-', '')
    provider['context'] = os.getcwd()

    build_provider(provider_name, provider, settings)
    deploy_provider(provider_name, provider, settings)

build_deploy_provider()

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
