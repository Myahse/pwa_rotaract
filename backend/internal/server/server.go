package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rotaract-civ/backend/internal/auth"
	"github.com/rotaract-civ/backend/internal/config"
	"github.com/rotaract-civ/backend/internal/database"
	"github.com/rotaract-civ/backend/internal/email"
	"github.com/rotaract-civ/backend/internal/handler"
	"github.com/rotaract-civ/backend/internal/middleware"
	"github.com/rotaract-civ/backend/internal/push"
	"github.com/rotaract-civ/backend/internal/repository"
	"github.com/rotaract-civ/backend/internal/scheduler"
	"github.com/rotaract-civ/backend/internal/service"
	"github.com/rotaract-civ/backend/internal/storage"
	"github.com/rotaract-civ/backend/internal/ws"
)

type Server struct {
	cfg               *config.Config
	logger            *slog.Logger
	store             *database.Store
	router            chi.Router
	BirthdayScheduler *scheduler.BirthdayScheduler
}

func New(cfg *config.Config, logger *slog.Logger, store *database.Store) (*Server, error) {
	s := &Server{
		cfg:    cfg,
		logger: logger,
		store:  store,
	}

	router, birthdayScheduler, err := s.buildRouter()
	if err != nil {
		return nil, err
	}
	s.router = router
	s.BirthdayScheduler = birthdayScheduler
	return s, nil
}

func (s *Server) Router() http.Handler {
	return s.router
}

func (s *Server) buildRouter() (chi.Router, *scheduler.BirthdayScheduler, error) {
	r := chi.NewRouter()

	fileStore, err := storage.NewLocalStore(s.cfg.UploadDir)
	if err != nil {
		return nil, nil, fmt.Errorf("init file storage: %w", err)
	}

	userRepo := repository.NewUserRepository(s.store.Pool)
	clubRepo := repository.NewClubRepository(s.store.Pool)
	commissionRepo := repository.NewCommissionRepository(s.store.Pool)
	chatRepo := repository.NewChatRepository(s.store.Pool)
	requestRepo := repository.NewAccessRequestRepository(s.store.Pool)
	clubRegistrationRepo := repository.NewClubRegistrationRepository(s.store.Pool)
	emailInviteRepo := repository.NewEmailInviteRepository(s.store.Pool)
	passwordResetRepo := repository.NewPasswordResetRepository(s.store.Pool)
	diaryRepo := repository.NewDiaryRepository(s.store.Pool)
	mandateRepo := repository.NewMandateRepository(s.store.Pool)
	pushSubRepo := repository.NewPushSubscriptionRepository(s.store.Pool)
	birthdayRepo := repository.NewBirthdayRepository(s.store.Pool)
	donationRepo := repository.NewDonationRepository(s.store.Pool)
	eventRepo := repository.NewPublicEventRepository(s.store.Pool)
	siteContentRepo := repository.NewSiteContentRepository(s.store.Pool)
	socialRepo := repository.NewSocialRepository(s.store.Pool)

	pushSender := push.NewSender(push.VAPIDConfig{
		PublicKey:  s.cfg.VAPID.PublicKey,
		PrivateKey: s.cfg.VAPID.PrivateKey,
		Subject:    s.cfg.VAPID.Subject,
	})

	tokenManager := auth.NewTokenManager(s.cfg.JWTSecret, s.cfg.JWTAccessTTL)
	mailer := email.NewClient(email.Config{
		Enabled:          s.cfg.Brevo.APIKey != "",
		BrevoAPIKey:      s.cfg.Brevo.APIKey,
		BrevoSenderEmail: s.cfg.Brevo.SenderEmail,
		BrevoSenderName:  s.cfg.Brevo.SenderName,
	}, s.logger)

	chatHub := ws.NewHub(s.logger)
	authService := service.NewAuthService(userRepo, tokenManager)
	profileService := service.NewProfileService(userRepo, fileStore, s.cfg.APIPublicURL, s.cfg.MaxAvatarSize)
	diaryService := service.NewDiaryService(diaryRepo, clubRepo, userRepo, profileService, fileStore, s.cfg.MaxAvatarSize)
	mandateService := service.NewMandateService(mandateRepo, diaryRepo, clubRepo, userRepo, profileService, fileStore, s.cfg.MaxAvatarSize)
	adminService := service.NewAdminService(userRepo, clubRepo, commissionRepo, chatRepo, diaryService, mandateService, profileService, fileStore, s.cfg.MaxAvatarSize)
	clubService := service.NewClubService(userRepo, clubRepo, commissionRepo, chatRepo)
	chatService := service.NewChatService(clubRepo, commissionRepo, chatRepo, profileService, chatHub)
	registrationService := service.NewRegistrationService(
		userRepo, clubRepo, chatRepo, requestRepo, emailInviteRepo, mailer, s.cfg.AppPublicURL, s.cfg.InviteTTL,
	)
	clubRegistrationService := service.NewClubRegistrationService(
		clubRegistrationRepo, userRepo, clubRepo, adminService, mailer, tokenManager,
		s.cfg.AppPublicURL, s.cfg.InviteTTL, s.cfg.GoogleClientID,
	)
	passwordResetService := service.NewPasswordResetService(userRepo, passwordResetRepo, mailer, s.cfg.AppPublicURL, s.cfg.PasswordResetTTL)
	donationService := service.NewDonationService(donationRepo, userRepo, fileStore, s.cfg.APIPublicURL)
	eventService := service.NewPublicEventService(eventRepo, fileStore, s.cfg.APIPublicURL, s.cfg.MaxAvatarSize)
	siteContentService := service.NewSiteContentService(siteContentRepo, fileStore, s.cfg.APIPublicURL, s.cfg.MaxAvatarSize)
	birthdayService, err := service.NewBirthdayService(
		birthdayRepo, pushSubRepo, profileService, pushSender,
		s.cfg.BirthdayTimezone, s.cfg.BirthdayNotifyHour, s.cfg.AppPublicURL, s.logger,
	)
	if err != nil {
		return nil, nil, err
	}
	birthdayScheduler := scheduler.NewBirthdayScheduler(birthdayService, s.logger)

	authHandler := handler.NewAuthHandler(authService)
	passwordResetHandler := handler.NewPasswordResetHandler(passwordResetService, s.logger)
	donationHandler := handler.NewDonationHandler(donationService)
	eventHandler := handler.NewPublicEventHandler(eventService)
	siteContentHandler := handler.NewSiteContentHandler(siteContentService)
	adminHandler := handler.NewAdminHandler(adminService)
	clubHandler := handler.NewClubHandler(clubService, clubRepo, profileService)
	chatHandler := handler.NewChatHandler(chatService, chatRepo, clubRepo)
	chatWSHandler := handler.NewChatWSHandler(chatService, chatHub, tokenManager)
	registrationHandler := handler.NewRegistrationHandler(registrationService, clubRepo, requestRepo, s.cfg.AppPublicURL)
	clubRegistrationHandler := handler.NewClubRegistrationHandler(clubRegistrationService)
	profileHandler := handler.NewProfileHandler(profileService, clubRepo, s.logger)
	diaryHandler := handler.NewDiaryHandler(diaryService, mandateService)
	birthdayHandler := handler.NewBirthdayHandler(birthdayService)
	socialService := service.NewSocialService(socialRepo, clubRepo, profileService, fileStore, 50<<20)
	socialHandler := handler.NewSocialHandler(socialService)
	internalHandler := handler.NewInternalHandler(birthdayScheduler, s.cfg.CronSecret)
	healthHandler := handler.NewHealthHandler(s.cfg, s.store)

	authenticate := middleware.Authenticate(tokenManager, userRepo)
	optionalAuth := middleware.OptionalAuthenticate(tokenManager, userRepo)

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logger(s.logger))
	r.Use(chimiddleware.Recoverer)
	corsOptions := cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "X-Cron-Secret"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}
	if s.cfg.CORSAllowAll {
		corsOptions.AllowOriginFunc = func(_ *http.Request, origin string) bool {
			return origin != ""
		}
	} else {
		corsOptions.AllowedOrigins = s.cfg.CORSOrigins
	}
	r.Use(cors.Handler(corsOptions))

	r.Get("/health", healthHandler.Health)
	r.Get("/ready", healthHandler.Ready)

	r.Handle("/api/v1/uploads/*", http.StripPrefix("/api/v1/uploads/", http.FileServer(http.Dir(fileStore.RootDir()))))

	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/status", healthHandler.Status)

		api.Post("/auth/login", authHandler.Login)
		api.Post("/auth/forgot-password", passwordResetHandler.Request)
		api.Post("/auth/reset-password", passwordResetHandler.Complete)
		api.Get("/invite/token/{token}", registrationHandler.PreviewEmailInvite)
		api.Get("/invite/{code}", registrationHandler.PreviewInvite)
		api.Post("/register", registrationHandler.Register)
		api.Post("/access-requests", registrationHandler.SubmitAccessRequest)
		api.Post("/donations", donationHandler.Submit)
		api.Get("/events", eventHandler.ListPublic)
		api.Get("/events/{eventID}", eventHandler.GetPublic)
		api.Get("/gallery", siteContentHandler.ListGalleryPublic)
		api.Get("/featured-postulant", siteContentHandler.GetFeaturedPostulantPublic)
		api.Post("/club-registration-requests", clubRegistrationHandler.Submit)
		api.Get("/club-registration/access/{token}", clubRegistrationHandler.PreviewAccess)
		api.Post("/club-registration/complete", clubRegistrationHandler.Complete)
		api.Get("/auth/google-client-id", clubRegistrationHandler.GoogleClientID)
		api.Get("/push/vapid-key", birthdayHandler.VAPIDPublicKey)
		api.Post("/internal/birthdays/run", internalHandler.RunBirthdays)
		api.Get("/ws/chat", chatWSHandler.ServeWS)

		// Public social read (guests welcome; optional auth enriches reactions/follows)
		api.Group(func(publicSocial chi.Router) {
			publicSocial.Use(optionalAuth)
			publicSocial.Get("/social/feed", socialHandler.ListFeed)
			publicSocial.Get("/social/posts/{postID}", socialHandler.GetPost)
			publicSocial.Get("/social/posts/{postID}/comments", socialHandler.ListComments)
			publicSocial.Get("/social/suggestions", socialHandler.Suggestions)
			publicSocial.Get("/social/users/{userID}", socialHandler.GetUserProfile)
			publicSocial.Get("/social/users/{userID}/posts", socialHandler.ListUserPosts)
			publicSocial.Get("/social/groups", socialHandler.ListGroups)
			publicSocial.Get("/social/groups/{groupID}", socialHandler.GetGroup)
			publicSocial.Get("/social/groups/{groupID}/members", socialHandler.ListGroupMembers)
		})

		api.Group(func(protected chi.Router) {
			protected.Use(authenticate)

			protected.Get("/auth/me", profileHandler.GetMe)
			protected.Get("/users/me/clubs", profileHandler.ListMyClubs)
			protected.Patch("/users/me", profileHandler.UpdateMe)
			protected.Post("/users/me/avatar", profileHandler.UploadAvatar)
			protected.Delete("/users/me/avatar", profileHandler.DeleteAvatar)

			protected.Post("/users/me/push-subscription", birthdayHandler.SavePushSubscription)
			protected.Delete("/users/me/push-subscription", birthdayHandler.DeletePushSubscription)
			protected.Get("/widgets/birthday", birthdayHandler.Widget)

			protected.Route("/admin", func(admin chi.Router) {
				admin.Use(middleware.RequireAdmin)
				admin.Post("/clubs", adminHandler.CreateClub)
				admin.Get("/clubs", adminHandler.ListClubs)
				admin.Patch("/clubs/{clubID}", adminHandler.UpdateClub)
				admin.Delete("/clubs/{clubID}", adminHandler.DeleteClub)
				admin.Post("/clubs/{clubID}/head", adminHandler.CreateClubHead)
				admin.Get("/access-requests", registrationHandler.ListAdminAccessRequests)
				admin.Post("/access-requests/{requestID}/approve", registrationHandler.ApproveAdminAccessRequest)
				admin.Post("/access-requests/{requestID}/reject", registrationHandler.RejectAccessRequest)
				admin.Get("/donations", donationHandler.List)
				admin.Post("/donations/{donationID}/received", donationHandler.MarkReceived)
				admin.Get("/events", eventHandler.ListAdmin)
				admin.Post("/events", eventHandler.Create)
				admin.Get("/events/{eventID}", eventHandler.GetAdmin)
				admin.Patch("/events/{eventID}", eventHandler.Update)
				admin.Delete("/events/{eventID}", eventHandler.Delete)
				admin.Post("/events/{eventID}/flyer", eventHandler.UploadFlyer)
				admin.Post("/events/{eventID}/images", eventHandler.AddImage)
				admin.Delete("/events/{eventID}/images/{imageID}", eventHandler.DeleteImage)

				admin.Get("/gallery", siteContentHandler.ListGalleryAdmin)
				admin.Post("/gallery", siteContentHandler.CreateGallery)
				admin.Patch("/gallery/{imageID}", siteContentHandler.UpdateGallery)
				admin.Post("/gallery/{imageID}/image", siteContentHandler.UploadGalleryImage)
				admin.Delete("/gallery/{imageID}", siteContentHandler.DeleteGallery)
				admin.Get("/featured-postulants", siteContentHandler.ListFeaturedPostulantsAdmin)
				admin.Post("/featured-postulants", siteContentHandler.CreateFeaturedPostulant)
				admin.Patch("/featured-postulants/{postulantID}", siteContentHandler.UpdateFeaturedPostulant)
				admin.Post("/featured-postulants/{postulantID}/flyer", siteContentHandler.UploadPostulantFlyer)
				admin.Delete("/featured-postulants/{postulantID}", siteContentHandler.DeleteFeaturedPostulant)

				admin.Get("/club-registration-requests", clubRegistrationHandler.List)
				admin.Get("/club-registration-requests/{requestID}", clubRegistrationHandler.Get)
				admin.Post("/club-registration-requests/{requestID}/approve", clubRegistrationHandler.Approve)
				admin.Post("/club-registration-requests/{requestID}/reject", clubRegistrationHandler.Reject)
			})

			protected.Route("/clubs/{clubID}", func(club chi.Router) {
				club.Get("/", clubHandler.GetClub)

				club.With(middleware.RequireClubPermission(clubRepo, "club.members.read")).Get("/members", clubHandler.ListMembers)
				club.With(middleware.RequireClubPermission(clubRepo, "club.members.read")).Get("/birthdays/today", birthdayHandler.ClubBirthdaysToday)
				club.With(middleware.RequireClubPermission(clubRepo, "club.members.read")).Get("/roles", clubHandler.ListRoles)
				club.With(middleware.RequireClubPermission(clubRepo, "club.commissions.read")).Get("/commissions", clubHandler.ListCommissions)
				club.With(middleware.RequireClubPermission(clubRepo, "chat.groups.read")).Get("/chat/groups", chatHandler.ListClubGroups)
				club.With(middleware.RequireClubPermission(clubRepo, "club.members.read")).Get("/invite", registrationHandler.GetClubInvite)
				club.With(middleware.RequireClubPermission(clubRepo, "club.invites.send")).Post("/email-invites", registrationHandler.SendEmailInvite)
				club.With(middleware.RequireClubPermission(clubRepo, "club.invites.send")).Get("/email-invites", registrationHandler.ListEmailInvites)
				club.With(middleware.RequireClubPermission(clubRepo, "club.invites.send")).Delete("/email-invites/{inviteID}", registrationHandler.RevokeEmailInvite)

				club.With(middleware.RequireClubPermission(clubRepo, "club.members.assign_role")).Post("/members/{userID}/roles", clubHandler.AssignRole)
				club.With(middleware.RequireClubPermission(clubRepo, "club.members.assign_role")).Delete("/members/{userID}/roles/{roleID}", clubHandler.UnassignRole)
				club.With(middleware.RequireClubPermission(clubRepo, "club.roles.manage")).Post("/roles", clubHandler.CreateRole)
				club.With(middleware.RequireClubPermission(clubRepo, "club.commissions.create")).Post("/commissions", clubHandler.CreateCommission)
				club.With(middleware.RequireClubPermission(clubRepo, "club.commissions.read")).Get("/commissions/{commissionID}", clubHandler.GetCommission)
				club.With(middleware.RequireClubPermission(clubRepo, "club.commissions.create")).Patch("/commissions/{commissionID}", clubHandler.UpdateCommission)
				club.With(middleware.RequireClubPermission(clubRepo, "club.commissions.create")).Delete("/commissions/{commissionID}", clubHandler.DeleteCommission)
				club.With(middleware.RequireClubPermission(clubRepo, "club.commissions.manage_members")).Post("/commissions/{commissionID}/members", clubHandler.AddCommissionMember)
				club.With(middleware.RequireClubPermission(clubRepo, "club.commissions.manage_members")).Delete("/commissions/{commissionID}/members/{userID}", clubHandler.RemoveCommissionMember)

				club.With(middleware.RequireClubPermission(clubRepo, "club.requests.review")).Get("/access-requests", registrationHandler.ListAccessRequests)
				club.With(middleware.RequireClubPermission(clubRepo, "club.requests.review")).Post("/access-requests/{requestID}/approve", registrationHandler.ApproveAccessRequest)
				club.With(middleware.RequireClubPermission(clubRepo, "club.requests.review")).Post("/access-requests/{requestID}/reject", registrationHandler.RejectAccessRequest)

				club.With(middleware.RequireClubPermission(clubRepo, "club.diary.read")).Get("/diary", diaryHandler.GetDiary)
				club.With(middleware.RequireClubPermission(clubRepo, "club.diary.manage")).Post("/diary", diaryHandler.CreateEntry)
				club.With(middleware.RequireClubPermission(clubRepo, "club.diary.manage")).Patch("/diary/{entryID}", diaryHandler.UpdateEntry)
				club.With(middleware.RequireClubPermission(clubRepo, "club.diary.manage")).Delete("/diary/{entryID}", diaryHandler.DeleteEntry)
				club.With(middleware.RequireClubPermission(clubRepo, "club.diary.manage")).Post("/diary/{entryID}/photo", diaryHandler.UploadPhoto)

				club.With(middleware.RequireClubPermission(clubRepo, "club.diary.manage")).Post("/mandates", diaryHandler.CreateMandate)
				club.With(middleware.RequireClubPermission(clubRepo, "club.diary.manage")).Delete("/mandates/{mandateID}", diaryHandler.DeleteMandate)
				club.With(middleware.RequireClubPermission(clubRepo, "club.diary.manage")).Post("/mandates/{mandateID}/assignments", diaryHandler.CreateMandateAssignment)
				club.With(middleware.RequireClubPermission(clubRepo, "club.diary.manage")).Delete("/mandates/{mandateID}/assignments/{assignmentID}", diaryHandler.DeleteMandateAssignment)

				club.With(middleware.RequireCommissionPresident(commissionRepo)).Post("/commissions/{commissionID}/chat/groups", chatHandler.CreateCommissionGroup)
			})

			protected.Get("/chat/groups/{groupID}", chatHandler.GetGroup)
			protected.Get("/chat/groups/{groupID}/messages", chatHandler.ListMessages)
			protected.Post("/chat/groups/{groupID}/messages", chatHandler.SendMessage)
			protected.Patch("/chat/groups/{groupID}/messages/{messageID}", chatHandler.UpdateMessage)
			protected.Delete("/chat/groups/{groupID}/messages/{messageID}", chatHandler.DeleteMessage)

			protected.Route("/social", func(social chi.Router) {
				social.Post("/posts", socialHandler.CreatePost)
				social.Post("/posts/{postID}/repost", socialHandler.Repost)
				social.Delete("/posts/{postID}", socialHandler.DeletePost)
				social.Post("/posts/{postID}/comments", socialHandler.CreateComment)
				social.Post("/posts/{postID}/reactions", socialHandler.React)
				social.Delete("/posts/{postID}/reactions", socialHandler.Unreact)
				social.Get("/following", socialHandler.ListFollowing)
				social.Post("/users/{userID}/follow", socialHandler.Follow)
				social.Delete("/users/{userID}/follow", socialHandler.Unfollow)

				social.Get("/friends", socialHandler.ListFriends)
				social.Get("/friends/requests", socialHandler.ListFriendRequests)
				social.Post("/friends/{userID}/request", socialHandler.RequestFriend)
				social.Post("/friends/{userID}/accept", socialHandler.AcceptFriend)
				social.Post("/friends/{userID}/decline", socialHandler.DeclineFriend)
				social.Delete("/friends/{userID}", socialHandler.RemoveFriend)

				social.Get("/conversations", socialHandler.ListConversations)
				social.Post("/conversations", socialHandler.OpenConversation)
				social.Get("/conversations/{conversationID}/messages", socialHandler.ListMessages)
				social.Post("/conversations/{conversationID}/messages", socialHandler.SendMessage)
				social.Post("/comments/{commentID}/share", socialHandler.ShareComment)

				social.Post("/groups", socialHandler.CreateGroup)
				social.Post("/groups/{groupID}/join", socialHandler.JoinGroup)
				social.Delete("/groups/{groupID}/leave", socialHandler.LeaveGroup)
				social.Get("/groups/{groupID}/requests", socialHandler.ListJoinRequests)
				social.Post("/groups/{groupID}/requests/{userID}/approve", socialHandler.ApproveJoinRequest)
				social.Post("/groups/{groupID}/requests/{userID}/reject", socialHandler.RejectJoinRequest)
				social.Get("/groups/{groupID}/messages", socialHandler.ListGroupMessages)
				social.Post("/groups/{groupID}/messages", socialHandler.SendGroupMessage)
			})
		})
	})

	return r, birthdayScheduler, nil
}
