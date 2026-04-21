package apperror

const (
	// 400
	ValidationMessageEmail           = "メールアドレスの形式が正しくありません"
	ValidationMessagePassword        = "パスワードの形式が正しくありません"
	ValidationMessageName            = "名前は必須です"
	ValidationMessageNameLength      = "名前は255文字以内である必要があります"
	ValidationMessageSku             = "SKUは必須です"
	ValidationMessageID              = "IDが正しくありません"
	ValidationMessageOrder           = "無効な注文IDです"
	ValidationMessageStatus          = "無効なステータスです"
	ValidationMessagePrice           = "価格は正の整数である必要があります"
	ValidationMessageRequest         = "リクエスト形式が正しくありません"
	ValidationMessageRole            = "無効なロールです"
	ValidationMessageCart            = "カートが空です"
	ValidationMessageCategory        = "カテゴリ名は必須です"
	ValidationMessageQty             = "在庫は1以上である必要があります"
	ValidationMessageGeneric         = "入力が不正です"
	ValidationMessageEssentialOrder  = "注文IDが必要です"
	ValidationMessageConflictedEmail = "このメールアドレスは既に登録されています"

	// 400
	BusinessLogicMessageRole = "自分自身のロールは変更できません"

	// 404
	NotFoundMessageProduct  = "商品が見つかりません"
	NotFoundMessageCart     = "カートが見つかりません"
	NotFoundMessageCartItem = "カートアイテムが見つかりません"
	NotFoundMessageUser     = "ユーザーが見つかりません"
	NotFoundMessageCategory = "カテゴリが見つかりません"
	NotFoundMessageOrder    = "注文が見つかりません"

	// 409
	ConflictMessageQty = "在庫不足です"
	ConflictMessageSku = "SKUが既に存在します"

	// 401
	UnauthorizedMessageAuth            = "認証が必要です"
	UnauthorizedMessageEmailOrPassword = "メールアドレスまたはパスワードが正しくありません"

	// 403
	ForbiddenMessageAdmin = "管理者権限が必要です"

	// 500
	InternalServerMessageCommon = "予期せぬエラーが発生しました"
)
