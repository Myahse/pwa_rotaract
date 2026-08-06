package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/rotaract-civ/backend/internal/domain"
)

type defaultCommission struct {
	Code        string
	Name        string
	Description string
}

var defaultClubCommissions = []defaultCommission{
	{
		Code:        "RP",
		Name:        "Relations Publiques",
		Description: "Communication du club : supports, flyers, templates et ressources graphiques.",
	},
	{
		Code:        "EFFECTIF",
		Name:        "Effectif",
		Description: "Gestion des membres : recrutement, suivi des adhésions et composition du club.",
	},
	{
		Code:        "COMM",
		Name:        "Communication",
		Description: "Coordination de la communication interne et externe, en lien avec Relations Publiques (RP).",
	},
	{
		Code:        "FONDATION",
		Name:        "Programmes de la Fondation",
		Description: "Organisation et suivi des programmes et projets de la Fondation Rotariaire.",
	},
	{
		Code:        "FINANCE",
		Name:        "Finance et Sponsoring",
		Description: "Gestion financière du club, budgets, trésorerie et recherche de sponsors.",
	},
	{
		Code:        "ADMIN",
		Name:        "Administration",
		Description: "Administration générale du club : procédures, documents officiels et organisation interne.",
	},
	{
		Code:        "RESTO",
		Name:        "Restauration",
		Description: "Organisation de la restauration lors des réunions, événements et activités du club.",
	},
	{
		Code:        "ACTIONS",
		Name:        "Actions",
		Description: "Planification et exécution des actions communautaires et projets de service du club.",
	},
	{
		Code:        "DEV_PRO",
		Name:        "Développement professionnel",
		Description: "Accompagnement des membres dans leur évolution professionnelle, réseau et opportunités de carrière.",
	},
	{
		Code:        "VIE_CLUB",
		Name:        "Vie du club",
		Description: "Animation de la vie associative : cohésion, événements internes et dynamique du club.",
	},
	{
		Code:        "ACTION_INTL",
		Name:        "Action internationale",
		Description: "Projets et actions à dimension internationale, partenariats et échanges inter-clubs.",
	},
	{
		Code:        "ACTION_PUBLIC",
		Name:        "Action d'intérêt public",
		Description: "Actions de service au bénéfice de la collectivité et de l'intérêt général.",
	},
	{
		Code:        "ACTION_INTER",
		Name:        "Action intérieure",
		Description: "Actions menées au sein du club et pour ses membres.",
	},
	{
		Code:        "INTERACT",
		Name:        "Interact",
		Description: "Relations et partenariats avec les clubs Interact.",
	},
	{
		Code:        "FORMATION",
		Name:        "Formation des membres",
		Description: "Organisation des formations, ateliers et montée en compétences des membres.",
	},
}

func (s *AdminService) seedDefaultCommissions(ctx context.Context, adminID, clubID uuid.UUID) error {
	for _, item := range defaultClubCommissions {
		code := item.Code
		desc := item.Description
		commission := &domain.Commission{
			ClubID:      clubID,
			Code:        &code,
			Name:        item.Name,
			Description: &desc,
			IsSystem:    true,
			CreatedBy:   &adminID,
		}
		if err := s.commissions.Create(ctx, commission); err != nil {
			return err
		}
	}
	return nil
}
