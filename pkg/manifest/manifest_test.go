package manifest

import (
	"io/ioutil"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/context"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestLoadManifest(t *testing.T) {
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/simple.porter.yaml", config.Name)

	m, err := LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	require.NotNil(t, m, "manifest was nil")
	require.Equal(t, m.Name, "hello", "manifest has incorrect name")
	require.Equal(t, m.Description, "An example Porter configuration", "manifest has incorrect description")
	require.Equal(t, m.Version, "0.1.0", "manifest has incorrect version")
	require.Equal(t, m.BundleTag, "getporter/porter-hello:v0.1.0", "manifest has incorrect bundle tag")

	assert.Equal(t, []MixinDeclaration{{Name: "exec"}}, m.Mixins, "expected manifest to declare the exec mixin")
	require.Len(t, m.Install, 1, "expected 1 install step")

	installStep := m.Install[0]
	description, _ := installStep.GetDescription()
	assert.NotNil(t, description, "expected the install description to be populated")

	mixin := installStep.GetMixinName()
	assert.Equal(t, "exec", mixin, "incorrect install step mixin used")

	require.Len(t, m.CustomActions, 1, "expected manifest to declare 1 custom action")
	require.Contains(t, m.CustomActions, "status", "expected manifest to declare a status action")

	statusStep := m.CustomActions["status"][0]
	description, _ = statusStep.GetDescription()
	assert.Equal(t, "Get World Status", description, "unexpected status step description")

	mixin = statusStep.GetMixinName()
	assert.Equal(t, "exec", mixin, "unexpected status step mixin")
}

func TestLoadManifest_DeprecatedFields(t *testing.T) {
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/porter-with-image.yaml", config.Name)

	m, err := LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "expected no error")
	require.NotNil(t, m, "manifest was nil")
}

func TestLoadManifestWithDependencies(t *testing.T) {
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/porter.yaml", config.Name)
	cxt.AddTestDirectory("testdata/bundles", "bundles")

	m, err := LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	assert.NotNil(t, m)
	assert.Equal(t, []MixinDeclaration{{Name: "exec"}}, m.Mixins)
	assert.Len(t, m.Install, 1)

	installStep := m.Install[0]
	description, _ := installStep.GetDescription()
	assert.NotNil(t, description)

	mixin := installStep.GetMixinName()
	assert.Equal(t, "exec", mixin)
}

func TestLoadManifestWithDependenciesInOrder(t *testing.T) {
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/porter-with-deps.yaml", config.Name)
	cxt.AddTestDirectory("testdata/bundles", "bundles")

	m, err := LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "could not load manifest")
	assert.NotNil(t, m)

	nginxDep := m.Dependencies[0]
	assert.Equal(t, "nginx", nginxDep.Name)
	assert.Equal(t, "localhost:5000/nginx:1.19", nginxDep.Tag)

	mysqlDep := m.Dependencies[1]
	assert.Equal(t, "mysql", mysqlDep.Name)
	assert.Equal(t, "getporter/azure-mysql:5.7", mysqlDep.Tag)
	assert.Len(t, mysqlDep.Parameters, 1)

}

func TestAction_Validate_RequireMixinDeclaration(t *testing.T) {
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/simple.porter.yaml", config.Name)

	m, err := LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	// Sabotage!
	m.Mixins = []MixinDeclaration{}

	err = m.Install.Validate(m)
	assert.EqualError(t, err, "mixin (exec) was not declared")
}

func TestAction_Validate_RequireMixinData(t *testing.T) {
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/simple.porter.yaml", config.Name)

	m, err := LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	// Sabotage!
	m.Install[0].Data = nil

	err = m.Install.Validate(m)
	assert.EqualError(t, err, "no mixin specified")
}

func TestAction_Validate_RequireSingleMixinData(t *testing.T) {
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/simple.porter.yaml", config.Name)

	m, err := LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	// Sabotage!
	m.Install[0].Data["rando-mixin"] = ""

	err = m.Install.Validate(m)
	assert.EqualError(t, err, "more than one mixin specified")
}

func TestManifest_Empty_Steps(t *testing.T) {
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/empty-steps.yaml", config.Name)

	_, err := LoadManifestFrom(cxt.Context, config.Name)
	assert.EqualError(t, err, "3 errors occurred:\n\t* validation of action \"install\" failed: found an empty step\n\t* validation of action \"uninstall\" failed: found an empty step\n\t* validation of action \"status\" failed: found an empty step\n\n")
}

func TestManifest_Validate_Dockerfile(t *testing.T) {
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/simple.porter.yaml", config.Name)

	m, err := LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	m.Dockerfile = "Dockerfile"

	err = m.Validate()

	assert.EqualError(t, err, "Dockerfile template cannot be named 'Dockerfile' because that is the filename generated during porter build")
}

func TestReadManifest_URL(t *testing.T) {
	cxt := context.NewTestContext(t)
	url := "https://raw.githubusercontent.com/deislabs/porter/v0.27.1/pkg/manifest/testdata/simple.porter.yaml"
	m, err := ReadManifest(cxt.Context, url)

	require.NoError(t, err)
	assert.Equal(t, "hello", m.Name)
}

func TestReadManifest_Validate_InvalidURL(t *testing.T) {
	cxt := context.NewTestContext(t)
	_, err := ReadManifest(cxt.Context, "http://fake-example-porter")

	assert.Error(t, err)
	assert.Regexp(t, "could not reach url http://fake-example-porter", err)
}

func TestReadManifest_File(t *testing.T) {
	cxt := context.NewTestContext(t)
	cxt.AddTestFile("testdata/simple.porter.yaml", config.Name)
	m, err := ReadManifest(cxt.Context, config.Name)

	require.NoError(t, err)
	assert.Equal(t, "hello", m.Name)
}

func TestSetDefault(t *testing.T) {
	t.Run("bundle docker tag set", func(t *testing.T) {
		m := Manifest{
			Version:   "1.2.3-beta.1",
			BundleTag: "getporter/mybun:v1.2.3",
		}
		err := m.SetDefaults()
		require.NoError(t, err)
		assert.Equal(t, "getporter/mybun:v1.2.3", m.BundleTag)
		assert.Equal(t, "getporter/mybun-installer:v1.2.3", m.Image)
	})

	t.Run("bundle docker tag not set", func(t *testing.T) {
		m := Manifest{
			Version:   "1.2.3-beta.1",
			BundleTag: "getporter/mybun",
		}
		err := m.SetDefaults()
		require.NoError(t, err)
		assert.Equal(t, "getporter/mybun:v1.2.3-beta.1", m.BundleTag)
		assert.Equal(t, "getporter/mybun-installer:v1.2.3-beta.1", m.Image)
	})

	t.Run("bundle tag includes registry with port", func(t *testing.T) {
		m := Manifest{
			Version:   "0.1.0",
			BundleTag: "localhost:5000/missing-invocation-image",
		}
		err := m.SetDefaults()
		require.NoError(t, err)
		assert.Equal(t, "localhost:5000/missing-invocation-image:v0.1.0", m.BundleTag)
		assert.Equal(t, "localhost:5000/missing-invocation-image-installer:v0.1.0", m.Image)
	})
}

func TestReadManifest_Validate_MissingFile(t *testing.T) {
	cxt := context.NewTestContext(t)
	_, err := ReadManifest(cxt.Context, "fake-porter.yaml")

	assert.EqualError(t, err, "the specified porter configuration file fake-porter.yaml does not exist")
}

func TestMixinDeclaration_UnmarshalYAML(t *testing.T) {
	cxt := context.NewTestContext(t)
	cxt.AddTestFile("testdata/mixin-with-config.yaml", config.Name)
	m, err := ReadManifest(cxt.Context, config.Name)

	require.NoError(t, err)
	assert.Len(t, m.Mixins, 2, "expected 2 mixins")
	assert.Equal(t, "exec", m.Mixins[0].Name)
	assert.Equal(t, "az", m.Mixins[1].Name)
	assert.Equal(t, map[interface{}]interface{}{"extensions": []interface{}{"iot"}}, m.Mixins[1].Config)
}

func TestMixinDeclaration_UnmarshalYAML_Invalid(t *testing.T) {
	cxt := context.NewTestContext(t)
	cxt.AddTestFile("testdata/mixin-with-bad-config.yaml", config.Name)
	_, err := ReadManifest(cxt.Context, config.Name)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "mixin declaration contained more than one mixin")
}

func TestCredentialsDefinition_UnmarshalYAML(t *testing.T) {
	assertAllCredentialsRequired := func(t *testing.T, creds CredentialDefinitions) {
		for _, cred := range creds {
			assert.EqualValuesf(t, true, cred.Required, "Credential: %s should be required", cred.Name)
		}
	}
	t.Run("all credentials in the generated manifest file are required", func(t *testing.T) {
		cxt := context.NewTestContext(t)
		cxt.AddTestFile("testdata/with-credentials.yaml", config.Name)
		m, err := ReadManifest(cxt.Context, config.Name)
		require.NoError(t, err)
		assertAllCredentialsRequired(t, m.Credentials)
	})
}

func TestMixinDeclaration_MarshalYAML(t *testing.T) {
	m := struct {
		Mixins []MixinDeclaration
	}{
		[]MixinDeclaration{
			{Name: "exec"},
			{Name: "az", Config: map[interface{}]interface{}{"extensions": []interface{}{"iot"}}},
		},
	}

	gotYaml, err := yaml.Marshal(m)
	require.NoError(t, err, "could not marshal data")

	wantYaml, err := ioutil.ReadFile("testdata/mixin-with-config.yaml")
	require.NoError(t, err, "could not read testdata")

	assert.Equal(t, string(wantYaml), string(gotYaml))
}

func TestValidateParameterDefinition(t *testing.T) {
	pd := ParameterDefinition{
		Name: "myparam",
		Schema: definition.Schema{
			Type: "file",
		},
	}

	pd.Destination = Location{}

	err := pd.Validate()
	assert.EqualError(t, err, `1 error occurred:
	* no destination path supplied for parameter myparam

`)

	pd.Destination.Path = "/path/to/file"

	err = pd.Validate()
	assert.NoError(t, err)
}

func TestValidateOutputDefinition(t *testing.T) {
	od := OutputDefinition{
		Name: "myoutput",
		Schema: definition.Schema{
			Type: "file",
		},
	}

	err := od.Validate()
	assert.EqualError(t, err, `1 error occurred:
	* no path supplied for output myoutput

`)

	od.Path = "/path/to/file"

	err = od.Validate()
	assert.NoError(t, err)
}

func TestValidateImageMap(t *testing.T) {
	t.Run("with both valid image digest and valid repository format", func(t *testing.T) {
		mi := MappedImage{
			Repository: "getporter/myserver",
			Digest:     "sha256:8f1133d81f1b078c865cdb11d17d1ff15f55c449d3eecca50190eed0f5e5e26f",
		}

		err := mi.Validate()
		// No error should be returned
		assert.NoError(t, err)
	})

	t.Run("with no image digest supplied and valid repository format", func(t *testing.T) {
		mi := MappedImage{
			Repository: "getporter/myserver",
		}

		err := mi.Validate()
		// No error should be returned
		assert.NoError(t, err)
	})

	t.Run("with valid image digest but invalid repository format", func(t *testing.T) {
		mi := MappedImage{
			Repository: "getporter//myserver//",
			Digest:     "sha256:8f1133d81f1b078c865cdb11d17d1ff15f55c449d3eecca50190eed0f5e5e26f",
		}

		err := mi.Validate()
		assert.Error(t, err)
	})

	t.Run("with invalid image digest format", func(t *testing.T) {
		mi := MappedImage{
			Repository: "getporter/myserver",
			Digest:     "abc123",
		}

		err := mi.Validate()
		assert.Error(t, err)
	})
}

func TestLoadManifestWithCustomData(t *testing.T) {
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/porter.yaml", config.Name)

	m, err := LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	assert.NotNil(t, m)
	assert.Equal(t, map[string]interface{}{"foo": "bar"}, m.Custom)

	custom := m.Custom
	fooCustomField, _ := custom["foo"]
	assert.Equal(t, "bar", fooCustomField)
}

func TestLoadManifestWithRequiredExtensions(t *testing.T) {
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/porter.yaml", config.Name)

	m, err := LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	expected := []RequiredExtension{
		RequiredExtension{
			Name: "requiredExtension1",
		},
		RequiredExtension{
			Name: "requiredExtension2",
			Config: map[string]interface{}{
				"config": true,
			},
		},
	}

	assert.NotNil(t, m)
	assert.Equal(t, expected, m.Required)
}

func TestReadManifest_WithTemplateVariables(t *testing.T) {
	cxt := context.NewTestContext(t)
	cxt.AddTestFile("testdata/porter-with-templating.yaml", config.Name)
	m, err := ReadManifest(cxt.Context, config.Name)
	require.NoError(t, err, "ReadManifest failed")
	wantVars := []string{"bundle.dependencies.mysql.outputs.mysql-password", "bundle.outputs.msg", "bundle.outputs.name"}
	assert.Equal(t, wantVars, m.TemplateVariables)
}

func TestManifest_GetTemplatedOutputs(t *testing.T) {
	cxt := context.NewTestContext(t)
	cxt.AddTestFile("testdata/porter-with-templating.yaml", config.Name)
	m, err := ReadManifest(cxt.Context, config.Name)
	require.NoError(t, err, "ReadManifest failed")

	outputs := m.GetTemplatedOutputs()

	require.Len(t, outputs, 1)
	assert.Equal(t, "msg", outputs["msg"].Name)
}

func TestManifest_GetTemplatedDependencyOutputs(t *testing.T) {
	cxt := context.NewTestContext(t)
	cxt.AddTestFile("testdata/porter-with-templating.yaml", config.Name)
	m, err := ReadManifest(cxt.Context, config.Name)
	require.NoError(t, err, "ReadManifest failed")

	outputs := m.GetTemplatedDependencyOutputs()

	require.Len(t, outputs, 1)
	ref := outputs["mysql.mysql-password"]
	assert.Equal(t, "mysql", ref.Dependency)
	assert.Equal(t, "mysql-password", ref.Output)
}

func TestParamToEnvVar(t *testing.T) {
	testcases := []struct {
		name      string
		paramName string
		envName   string
	}{
		{"no special characters", "myparam", "MYPARAM"},
		{"dash", "my-param", "MY_PARAM"},
		{"period", "my.param", "MY_PARAM"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParamToEnvVar(tc.paramName)
			assert.Equal(t, tc.envName, got)
		})
	}
}

func TestParameterDefinition_UpdateApplyTo(t *testing.T) {
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/simple.porter.yaml", config.Name)

	m, err := LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	testcases := []struct {
		name         string
		defaultValue string
		applyTo      []string
		source       ParameterSource
		wantApplyTo  []string
	}{
		{"no source", "", nil, ParameterSource{}, nil},
		{"has default", "myparam", nil, ParameterSource{Output: "myoutput"}, nil},
		{"has applyTo", "", []string{"status"}, ParameterSource{Output: "myoutput"}, []string{"status"}},
		{"no default, no applyTo", "", nil, ParameterSource{Output: "myoutput"}, []string{"status", "uninstall"}},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			pd := ParameterDefinition{
				Name: "myparam",
				Schema: definition.Schema{
					Type: "file",
				},
				Source:  tc.source,
				ApplyTo: tc.applyTo,
			}

			if tc.defaultValue != "" {
				pd.Schema.Default = tc.defaultValue
			}

			pd.UpdateApplyTo(m)
			require.Equal(t, tc.wantApplyTo, pd.ApplyTo)
		})
	}
}
