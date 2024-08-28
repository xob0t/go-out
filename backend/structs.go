package backend

type MergeSettings struct {
	ExifTags              ExifTags `json:"exifTags" koanf:"exif_tags"`
	IgnoreMinorErrors     bool     `json:"ignoreMinorErrors" koanf:"ignore_minor_errors"`
	EditedSuffix          string   `json:"editedSuffix" koanf:"edited_suffix"`
	TimezoneOffset        string   `json:"timezoneOffset" koanf:"timezone_offset"`
	InferTimezoneFromGPS  bool     `json:"inferTimezoneFromGPS" koanf:"infer_timezone_from_GPS"`
	OverwriteExistingTags bool     `json:"overwriteExistingTags" koanf:"overwrite_existing_tags"`
}

type ExifTags struct {
	Title       bool `json:"title" koanf:"title"`
	Description bool `json:"description" koanf:"description"`
	DateTaken   bool `json:"dateTaken" koanf:"date_taken"`
	URL         bool `json:"URL" koanf:"URL"`
	GPS         bool `json:"GPS" koanf:"GPS"`
}

type GlobalSettingsT struct {
	MergeSettings MergeSettings `json:"mergeSettings" koanf:"merge_settings"`
	Window        struct {
		Saved       bool `json:"saved" koanf:"saved"`
		IsMaximised bool `json:"isMaximised" koanf:"is_maximised"`
		SizeW       int  `json:"sizeW" koanf:"size_w"`
		SizeH       int  `json:"sizeH" koanf:"size_h"`
		PosX        int  `json:"posX" koanf:"pos_x"`
		PosY        int  `json:"posY" koanf:"pos_y"`
	} `json:"window" koanf:"window"`
}
