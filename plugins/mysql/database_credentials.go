package mysql

import (
	"context"
	"fmt"

	"github.com/1Password/shell-plugins/sdk"
	"github.com/1Password/shell-plugins/sdk/importer"
	"github.com/1Password/shell-plugins/sdk/provision"
	"github.com/1Password/shell-plugins/sdk/schema"
	"github.com/1Password/shell-plugins/sdk/schema/credname"
	"github.com/1Password/shell-plugins/sdk/schema/fieldname"
)

func DatabaseCredentials() schema.CredentialType {
	return schema.CredentialType{
		Name:    credname.DatabaseCredentials,
		DocsURL: sdk.URL("https://dev.mysql.com/doc/refman/en/connecting.html"),
		Fields: []schema.CredentialField{
			{
				Name:                fieldname.Host,
				MarkdownDescription: "MySQL host to connect to.",
				Optional:            true,
			},
			{
				Name:                fieldname.Port,
				MarkdownDescription: "Port used to connect to MySQL.",
				Optional:            true,
			},
			{
				Name:                fieldname.User,
				MarkdownDescription: "MySQL user to authenticate as.",
				Optional:            true,
			},
			{
				Name:                fieldname.Password,
				MarkdownDescription: "Password used to authenticate to MySQL.",
				Secret:              true,
			},
			{
				Name:                fieldname.Database,
				MarkdownDescription: "Database name to connect to.",
				Optional:            true,
			},
		},
		Provisioner: provision.TempFile(mysqlConfig, provision.Filename("my.cnf"), provision.SetPathAsArg("--defaults-file")),
		Importer: importer.TryAll(
			TryMySQLConfigFile("/etc/my.cnf"),
			TryMySQLConfigFile("/etc/mysql/my.cnf"),
			TryMySQLConfigFile("~/.my.cnf"),
			TryMySQLConfigFile("~/.mylogin.cnf"),
		),
	}
}

func mysqlConfig(in sdk.ProvisionInput) ([]byte, error) {
	config := map[string]string{
		"host": "127.0.0.1", // Default host
		"port": "3306",      // Default port
	}

	if user, ok := in.ItemFields[fieldname.User]; ok {
		config["user"] = user
	}

	if password, ok := in.ItemFields[fieldname.Password]; ok {
		config["password"] = password
	}

	if host, ok := in.ItemFields[fieldname.Host]; ok {
		config["host"] = host
	}

	if port, ok := in.ItemFields[fieldname.Port]; ok {
		config["port"] = port
	}

	if database, ok := in.ItemFields[fieldname.Database]; ok {
		config["database"] = database
	}

	content := "[client]\n"
	for key, value := range config {
		configLine := fmt.Sprintf("%s=%s\n", key, value)
		content += configLine
	}

	return []byte(content), nil
}

func TryMySQLConfigFile(path string) sdk.Importer {
	return importer.TryFile(path, func(ctx context.Context, contents importer.FileContents, in sdk.ImportInput, out *sdk.ImportAttempt) {
		credentialsFile, err := contents.ToINI()
		if err != nil {
			out.AddError(err)
			return
		}

		var fields []sdk.ImportCandidateField
		for _, section := range credentialsFile.Sections() {
			if section.HasKey("user") && section.Key("user").Value() != "" {
				fields = append(fields, sdk.ImportCandidateField{
					Field: "user",
					Value: section.Key("user").Value(),
				})
			}

			if section.HasKey("password") && section.Key("password").Value() != "" {
				fields = append(fields, sdk.ImportCandidateField{
					Field: "password",
					Value: section.Key("password").Value(),
				})
			}

			if section.HasKey("database") && section.Key("database").Value() != "" {
				fields = append(fields, sdk.ImportCandidateField{
					Field: "database",
					Value: section.Key("database").Value(),
				})
			}

			if section.HasKey("host") && section.Key("host").Value() != "" {
				fields = append(fields, sdk.ImportCandidateField{
					Field: "host",
					Value: section.Key("host").Value(),
				})
			}

			if section.HasKey("port") && section.Key("port").Value() != "" {
				fields = append(fields, sdk.ImportCandidateField{
					Field: "port",
					Value: section.Key("port").Value(),
				})
			}
		}

		out.AddCandidate(sdk.ImportCandidate{
			Fields: fields,
		})
	})
}
