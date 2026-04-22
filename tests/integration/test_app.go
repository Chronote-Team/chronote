package integration

import appplatform "chronote-refactor/internal/platform/app"

type TestApp = appplatform.App

func NewTestApp() (*TestApp, error) {
	return appplatform.NewTestApp()
}
