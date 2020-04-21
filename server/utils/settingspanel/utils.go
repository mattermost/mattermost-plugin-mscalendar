package settingspanel

import "github.com/mattermost/mattermost-server/v5/model"

func stringsToOptions(in []string) []*model.PostActionOptions {
	out := []*model.PostActionOptions{}
	for _, o := range in {
		out = append(out, &model.PostActionOptions{
			Text:  o,
			Value: o,
		})
	}
	return out
}
