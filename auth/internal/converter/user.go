package converter

import (
	"inventory-go/auth/internal/model"
	authV1 "inventory-go/shared/pkg/proto/auth/v1"
)

func NotificationMethodsFromProto(in []*authV1.NotificationMethod) []model.NotificationMethod {
	out := make([]model.NotificationMethod, 0, len(in))
	for _, m := range in {
		out = append(out, model.NotificationMethod{
			ProviderName: m.GetProviderName(),
			Target:       m.GetTarget(),
		})
	}

	return out
}

func NotificationMethodsToProto(in []model.NotificationMethod) []*authV1.NotificationMethod {
	out := make([]*authV1.NotificationMethod, 0, len(in))
	for _, m := range in {
		out = append(out, &authV1.NotificationMethod{
			ProviderName: m.ProviderName,
			Target:       m.Target,
		})
	}

	return out
}
