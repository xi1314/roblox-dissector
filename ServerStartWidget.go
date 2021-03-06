package main

import (
	"fmt"

	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/gskartwii/roblox-dissector/datamodel"
	"github.com/robloxapi/rbxfile"
	"github.com/therecipe/qt/widgets"
)

var noLocalDefaults = map[string](map[string]rbxfile.Value){
	"StarterGui": map[string]rbxfile.Value{
		"Archivable":            rbxfile.ValueBool(true),
		"Name":                  rbxfile.ValueString("StarterGui"),
		"ResetPlayerGuiOnSpawn": rbxfile.ValueBool(false),
		"RobloxLocked":          rbxfile.ValueBool(false),
		// TODO: Set token ID correctly here_
		"ScreenOrientation":  datamodel.ValueToken{Value: 0},
		"ShowDevelopmentGui": rbxfile.ValueBool(true),
		"Tags":               rbxfile.ValueBinaryString(""),
	},
	"Workspace": map[string]rbxfile.Value{
		"Archivable": rbxfile.ValueBool(true),
		// TODO: Set token ID correctly here_
		"AutoJointsMode":             datamodel.ValueToken{Value: 0},
		"CollisionGroups":            rbxfile.ValueString(""),
		"ExpSolverEnabled_Replicate": rbxfile.ValueBool(true),
		"ExplicitAutoJoints":         rbxfile.ValueBool(true),
		"FallenPartsDestroyHeight":   rbxfile.ValueFloat(-500.0),
		"FilteringEnabled":           rbxfile.ValueBool(true),
		"Gravity":                    rbxfile.ValueFloat(196.2),
		"ModelInPrimary":             rbxfile.ValueCFrame{},
		"Name":                       rbxfile.ValueString("Workspace"),
		"PrimaryPart":                datamodel.ValueReference{Instance: nil, Reference: datamodel.NullReference},
		"RobloxLocked":               rbxfile.ValueBool(false),
		"StreamingMinRadius":         rbxfile.ValueInt(0),
		"StreamingTargetRadius":      rbxfile.ValueInt(0),
		"Tags":                       rbxfile.ValueBinaryString(""),
		"TerrainWeldsFixed":          rbxfile.ValueBool(true),
	},
	"StarterPack": map[string]rbxfile.Value{
		"Archivable":   rbxfile.ValueBool(true),
		"Name":         rbxfile.ValueString("StarterPack"),
		"RobloxLocked": rbxfile.ValueBool(false),
		"Tags":         rbxfile.ValueBinaryString(""),
	},
	"TeleportService": map[string]rbxfile.Value{
		"Archivable":   rbxfile.ValueBool(true),
		"Name":         rbxfile.ValueString("Teleport Service"), // intentional
		"RobloxLocked": rbxfile.ValueBool(false),
		"Tags":         rbxfile.ValueBinaryString(""),
	},
	"LocalizationService": map[string]rbxfile.Value{
		"Archivable":           rbxfile.ValueBool(true),
		"IsTextScraperRunning": rbxfile.ValueBool(false),
		"LocaleManifest":       rbxfile.ValueString("en-us"),
		"Name":                 rbxfile.ValueString("LocalizationService"),
		"RobloxLocked":         rbxfile.ValueBool(false),
		"Tags":                 rbxfile.ValueBinaryString(""),
		"WebTableContents":     rbxfile.ValueString(""),
	},
}

// normalizeTypes changes the types of instances from binary format types to network types
func normalizeTypes(children []*datamodel.Instance, schema *peer.StaticSchema) {
	for _, instance := range children {
		defaultValues, ok := noLocalDefaults[instance.ClassName]
		if ok {
			for _, prop := range schema.Instances[schema.ClassesByName[instance.ClassName]].Properties {
				if _, ok = instance.Properties[prop.Name]; !ok {
					println("Adding missing default value", instance.ClassName, prop.Name)
					instance.Properties[prop.Name] = defaultValues[prop.Name]
				}
			}
		}

		for name, prop := range instance.Properties {
			id, ok := schema.PropertiesByName[instance.ClassName+"."+name]
			if !ok {
				fmt.Printf("Warning: %s.%s doesn't exist in schema! Stripping this property.\n", instance.ClassName, name)
				delete(instance.Properties, name)
				continue
			}
			propSchema := schema.Properties[id]
			switch propSchema.Type {
			case peer.PROP_TYPE_PROTECTEDSTRING_0,
				peer.PROP_TYPE_PROTECTEDSTRING_1,
				peer.PROP_TYPE_PROTECTEDSTRING_2,
				peer.PROP_TYPE_PROTECTEDSTRING_3:
				// This type may be encoded correctly depending on the format
				if _, ok = prop.(rbxfile.ValueString); ok {
					instance.Properties[name] = rbxfile.ValueProtectedString(prop.(rbxfile.ValueString))
				}
			case peer.PROP_TYPE_CONTENT:
				// This type may be encoded correctly depending on the format
				if _, ok = prop.(rbxfile.ValueString); ok {
					instance.Properties[name] = rbxfile.ValueContent(prop.(rbxfile.ValueString))
				}
			case peer.PROP_TYPE_ENUM:
				instance.Properties[name] = datamodel.ValueToken{ID: propSchema.EnumID, Value: prop.(datamodel.ValueToken).Value}
			case peer.PROP_TYPE_BINARYSTRING:
				// This type may be encoded correctly depending on the format
				if _, ok = prop.(rbxfile.ValueString); ok {
					instance.Properties[name] = rbxfile.ValueBinaryString(prop.(rbxfile.ValueString))
				}
			case peer.PROP_TYPE_COLOR3UINT8:
				if _, ok = prop.(rbxfile.ValueColor3); ok {
					propc3 := prop.(rbxfile.ValueColor3)
					instance.Properties[name] = rbxfile.ValueColor3uint8{R: uint8(propc3.R * 255), G: uint8(propc3.G * 255), B: uint8(propc3.B * 255)}
				}
			case peer.PROP_TYPE_BRICKCOLOR:
				if _, ok = prop.(rbxfile.ValueInt); ok {
					instance.Properties[name] = rbxfile.ValueBrickColor(prop.(rbxfile.ValueInt))
				}
			}
		}
		normalizeTypes(instance.Children, schema)
	}
}

func NewServerStartWidget(parent widgets.QWidget_ITF, settings *ServerSettings, callback func(*ServerSettings)) {
	window := widgets.NewQWidget(parent, 1)
	window.SetWindowTitle("Start server...")
	layout := widgets.NewQVBoxLayout()

	rbxlLabel := NewQLabelF("RBXLX location:")
	rbxlTextBox := widgets.NewQLineEdit2(settings.RBXLLocation, nil)
	browseButton := widgets.NewQPushButton2("Browse...", nil)
	browseButton.ConnectPressed(func() {
		rbxlTextBox.SetText(widgets.QFileDialog_GetOpenFileName(window, "Find place...", "", "RBXLX files (*.rbxlx)", "", 0))
	})
	layout.AddWidget(rbxlLabel, 0, 0)
	layout.AddWidget(rbxlTextBox, 0, 0)
	layout.AddWidget(browseButton, 0, 0)

	enumLabel := NewQLabelF("Enum schema location:")
	enumTextBox := widgets.NewQLineEdit2(settings.EnumSchemaLocation, nil)
	browseButton = widgets.NewQPushButton2("Browse...", nil)
	browseButton.ConnectPressed(func() {
		enumTextBox.SetText(widgets.QFileDialog_GetOpenFileName(window, "Find enum schema...", "", "Text files (*.txt)", "", 0))
	})
	layout.AddWidget(enumLabel, 0, 0)
	layout.AddWidget(enumTextBox, 0, 0)
	layout.AddWidget(browseButton, 0, 0)

	instanceLabel := NewQLabelF("Instance schema location:")
	instanceTextBox := widgets.NewQLineEdit2(settings.InstanceSchemaLocation, nil)
	browseButton = widgets.NewQPushButton2("Browse...", nil)
	browseButton.ConnectPressed(func() {
		instanceTextBox.SetText(widgets.QFileDialog_GetOpenFileName(window, "Find instance schema...", "", "Text files (*.txt)", "", 0))
	})
	layout.AddWidget(instanceLabel, 0, 0)
	layout.AddWidget(instanceTextBox, 0, 0)
	layout.AddWidget(browseButton, 0, 0)

	// HACK: convenience
	if settings.Port == "" {
		settings.Port = "53640"
	}
	portLabel := NewQLabelF("Port number:")
	port := widgets.NewQLineEdit2(settings.Port, nil)
	layout.AddWidget(portLabel, 0, 0)
	layout.AddWidget(port, 0, 0)

	startButton := widgets.NewQPushButton2("Start", nil)
	startButton.ConnectPressed(func() {
		window.Destroy(true, true)
		settings.Port = port.Text()
		settings.EnumSchemaLocation = enumTextBox.Text()
		settings.InstanceSchemaLocation = instanceTextBox.Text()
		settings.RBXLLocation = rbxlTextBox.Text()
		callback(settings)
	})
	layout.AddWidget(startButton, 0, 0)

	window.SetLayout(layout)
	window.Show()
}
