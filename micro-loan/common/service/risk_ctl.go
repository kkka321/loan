package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"encoding/json"
	"micro-loan/common/dao"
	"micro-loan/common/i18n"
	"micro-loan/common/lib/sms/areacode"
	"micro-loan/common/models"
	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/pkg/monitor"
	"micro-loan/common/pkg/schema_task"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/pkg/ticket"
	"micro-loan/common/thirdparty/advance"
	"micro-loan/common/thirdparty/voip"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type ModuleSN int

const (
	ReasonSq string = "ReasionSq"
)

const (
	QuestionStatusUndefined int = 0
	QuestionStatusNormal    int = 1
	QuestionStatusAbnormal  int = 2
)

var questionStatusMap = map[int]string{
	QuestionStatusUndefined: "未定义",
	QuestionStatusNormal:    "正常",
	QuestionStatusAbnormal:  "异常",
}

const (
	ModuleIdentity   ModuleSN = 10 // 身份信息
	ModuleLoan       ModuleSN = 11 // 借款信息
	ModuleWork       ModuleSN = 12 // 工作信息
	ModuleContacts   ModuleSN = 13 // 联系人信息
	ModuleOther      ModuleSN = 14 // 其他信息
	ModuleFixed      ModuleSN = 15 // 固定问题
	ModuleReloan     ModuleSN = 16 // 复贷问题
	ModuleInfoReview ModuleSN = 18 // Info Review问题
)

type PhoneVerifyQuestionItem struct {
	QuestionSN int             // 问题编号
	Question   i18n.LangItem   // 问题
	InputType  string          // 输入input类型
	Field      string          // 所对应的字段名
	Reasons    []i18n.LangItem // 拒绝原因  目前只有infoReview用到
}

var moduleSNMap = map[int]ModuleSN{
	0: ModuleIdentity,
	1: ModuleLoan,
	//2: ModuleWork,
	//3: ModuleContacts,
	4: ModuleOther,
}

var reloanModuleSNMap = map[int]ModuleSN{
	1: ModuleLoan,
	2: ModuleWork,
	3: ModuleContacts,
	5: ModuleReloan,
}

var fieldMap = map[string]string{
	"OrderId":           "o.id",
	"AccountId":         "a.id",
	"Amount":            "o.amount",
	"Loan":              "o.loan",
	"Period":            "o.period",
	"RandomValue":       "o.random_value",
	"FixedRandom":       "o.fixed_random",
	"ApplyTime":         "o.apply_time",
	"RiskCtlFinishTime": "o.risk_ctl_finish_time",
	"PhoneVerifyTime":   "o.phone_verify_time",
}

var phoneVerifyQuestionConfig = map[ModuleSN]map[int]PhoneVerifyQuestionItem{
	ModuleIdentity: map[int]PhoneVerifyQuestionItem{
		10001: {
			QuestionSN: 10001,
			Question: i18n.LangItem{
				i18n.LangZhCN: "您的身份证注册地是哪里？",
				i18n.LangIdID: "Di mana pendaftaran ID Anda?",
			},
			InputType: "text",
			Field:     "identity",
		},
		10002: {
			QuestionSN: 10002,
			Question: i18n.LangItem{
				i18n.LangZhCN: "您的身份证是何时注册的？",
				i18n.LangIdID: "Kapan kartu ID Anda mendaftar?",
			},
			InputType: "text",
			Field:     "identity",
		},
		10003: {
			QuestionSN: 10003,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您身份证的照片是戴眼镜的吗？/您身份证的照片有戴头巾吗？/……`,
				i18n.LangIdID: `Apakah foto kacamata identitas Anda memakai kacamata? / Apakah Anda memakai sorban di foto ID Anda? / ......`,
			},
			InputType: "radio",
			Field:     "id_photo",
		},
		10004: {
			QuestionSN: 10004,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您是在家中/公司/……进行活体认证的吗？`,
				i18n.LangIdID: `Apakah Anda tinggal di rumah / perusahaan / outdoor / ... biometrics hidup?`,
			},
			InputType: "radio",
			Field:     "image_env",
		},
		10005: {
			QuestionSN: 10005,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您活体认证时有戴眼镜的吗？/您活体认证时有戴头巾吗？/……`,
				i18n.LangIdID: `Apakah Anda memakai kacamata / saputangan / ... ketika Anda hidup bersertifikat?`,
			},
			InputType: "radio",
			Field:     "image_env",
		},
		10006: {
			QuestionSN: 10006,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您的年龄是？`,
				i18n.LangIdID: `Berapa umurmu?`,
			},
			InputType: "radio",
			Field:     "age",
		},
		10007: {
			QuestionSN: 10007,
			Question: i18n.LangItem{
				i18n.LangZhCN: `声音能明显判断性别的话，基础话术部分即可确认。如不能判断：抱歉，请问该称呼您**先生/女士？`,
				i18n.LangIdID: `Jika suara dapat dengan jelas menentukan jenis kelamin, bagian pembicaraan dasar dapat dikonfirmasikan. Jika Anda tidak bisa menilai: Maaf, saya harus memanggil Anda Mr.`,
			},
			InputType: "radio",
			Field:     "gender",
		},
		10008: {
			QuestionSN: 10008,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您的身份证号码是？`,
				i18n.LangIdID: `Berapa nomor ID Anda?`,
			},
			InputType: "radio",
			Field:     "identity",
		},
	},

	ModuleLoan: map[int]PhoneVerifyQuestionItem{
		11001: {
			QuestionSN: 11001,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您申请的金额是多少？`,
				i18n.LangIdID: `Berapa jumlah yang Anda ajukan?`,
			},
			InputType: "radio",
			Field:     "loan",
		},
		11002: {
			QuestionSN: 11002,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您申请的天数是多少？`,
				i18n.LangIdID: `Berapa hari Anda mengajukan permohonan?`,
			},
			InputType: "radio",
			Field:     "period",
		},
		11003: {
			QuestionSN: 11003,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您是何时申请的？/您是10:00（可故意给出错误时间）申请借款的吗？`,
				i18n.LangIdID: `Kapan Anda mendaftar? / Apakah ini 10: 00 (apakah Anda dapat dengan sengaja memberikan waktu yang salah) untuk mengajukan pinjaman?`,
			},
			InputType: "radio",
			Field:     "apply_time",
		},
		11004: {
			QuestionSN: 11004,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您是通过什么型号的手机申请的？`,
				i18n.LangIdID: `Jenis ponsel apa yang Anda lamar?`,
			},
			InputType: "radio",
			Field:     "brand",
		},
		11005: {
			QuestionSN: 11005,
			Question: i18n.LangItem{
				i18n.LangZhCN: `这笔贷款的用途是什么？`,
				i18n.LangIdID: `Apa gunanya Anda menggunakan pinjaman ini? (seperti membeli rumah / saham / perjudian / obat-obatan dan tujuan berbahaya lainnya, kemudian langsung menolak masalah nuklir lainnya)`,
			},
			InputType: "text",
			Field:     "",
		},
		11006: {
			QuestionSN: 11006,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您是从何处了解到我们的？`,
				i18n.LangIdID: `Darimana kamu belajar dari Rupiah Cepat?`,
			},
			InputType: "text",
			Field:     "",
		},
		11007: {
			QuestionSN: 11007,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您是否有从其他机构借款？已经结清了吗？`,
				i18n.LangIdID: `Apakah Anda meminjam dari institusi lain? Sudah siap?`,
			},
			InputType: "text",
			Field:     "",
		},
	},

	ModuleWork: map[int]PhoneVerifyQuestionItem{
		12001: {
			QuestionSN: 12001,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您目前在什么单位工作？`,
				i18n.LangIdID: `Anda bekerja di unit apa?`,
			},
			InputType: "radio",
			Field:     "company_name",
		},
		12002: {
			QuestionSN: 12002,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您是全职吗？/您是兼职吗？`,
				i18n.LangIdID: `Apakah kamu bekerja penuh waktu? / Apakah kamu paruh waktu?`,
			},
			InputType: "radio",
			Field:     "job_type",
		},
		12003: {
			QuestionSN: 12003,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您单位的地址是？`,
				i18n.LangIdID: `Apa alamat organisasi Anda?`,
			},
			InputType: "radio",
			Field:     "company_address",
		},
		12004: {
			QuestionSN: 12004,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您目前的月收入是？`,
				i18n.LangIdID: `Berapa penghasilan bulanan Anda saat ini?`,
			},
			InputType: "radio",
			Field:     "monthly_income",
		},
		12005: {
			QuestionSN: 12005,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您单位一般是什么时间发放薪水？`,
				i18n.LangIdID: `Berapa waktu umum Anda untuk membayar?`,
			},
			InputType: "text",
			Field:     "",
		},
		12006: {
			QuestionSN: 12006,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您单位现有有多少员工？`,
				i18n.LangIdID: `Berapa banyak karyawan yang ada di organisasi Anda?`,
			},
			InputType: "text",
			Field:     "",
		},
		12007: {
			QuestionSN: 12007,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您一般都是如何去上班？`,
				i18n.LangIdID: `Bagaimana Anda biasanya pergi bekerja? (Untuk referensi, silakan merujuk ke rute lalu lintas dari alamat rumah pelanggan ke alamat unit.)`,
			},
			InputType: "text",
			Field:     "",
		},
	},

	ModuleContacts: map[int]PhoneVerifyQuestionItem{
		13001: {
			QuestionSN: 13001,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您认识***（联系人1）吗？`,
				i18n.LangIdID: `Apakah Anda tahu *** (Hubungi 1/2)?`,
			},
			InputType: "radio",
			Field:     "contact1_name",
		},
		13002: {
			QuestionSN: 13002,
			Question: i18n.LangItem{
				i18n.LangZhCN: `***（联系人1）和您是什么关系？`,
				i18n.LangIdID: `Apa hubungan antara *** (Contact 1/2) dan Anda?`,
			},
			InputType: "radio",
			Field:     "relationship1",
		},
		13003: {
			QuestionSN: 13003,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您和***（联系人1）经常电话联系吗？`,
				i18n.LangIdID: `Apakah Anda dalam kontak telepon tetap dengan *** (Hubungi 1)?`,
			},
			InputType: "radio",
			Field:     "es_contact1",
		},
		//13004: PhoneVerifyQuestionItem{
		//	QuestionSN: 13004,
		//	Question:   `您和***（联系人1）一般何时联系吗？`,
		//	InputType:  "radio",
		//	Field:      "es_contact1_time",
		//},
		13005: {
			QuestionSN: 13005,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您知道***（联系人1）的电话号码吗？`,
				i18n.LangIdID: `Apakah Anda tahu nomor telepon dari *** (Hubungi 1)? (Ini pertanyaan terbuka dan mengharuskan pemohon untuk memberikan nomor lengkap)`,
			},
			InputType: "radio",
			Field:     "contact1",
		},
		13006: {
			QuestionSN: 13006,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您认识***（联系人2）吗？`,
				i18n.LangIdID: `Apakah kamu tahu *** (Kontak 2)?`,
			},
			InputType: "radio",
			Field:     "contact2_name",
		},
		13007: {
			QuestionSN: 13007,
			Question: i18n.LangItem{
				i18n.LangZhCN: `***（联系人2）和您是什么关系？`,
				i18n.LangIdID: `Apakah Anda tahu *** (Hubungi 2)?`,
			},
			InputType: "radio",
			Field:     "relationship2",
		},
		13008: {
			QuestionSN: 13008,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您和***（联系人2）经常电话联系吗？`,
				i18n.LangIdID: `Apakah Anda dalam kontak telepon tetap dengan *** (Hubungi 2)?`,
			},
			InputType: "radio",
			Field:     "es_contact2",
		},
		//13009: PhoneVerifyQuestionItem{
		//	QuestionSN: 13009,
		//	Question:   `您和***（联系人2）一般何时联系吗？`,
		//	InputType:  "radio",
		//	Field:      "es_contact2_time",
		//},
		13010: {
			QuestionSN: 13010,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您知道***（联系人2）的电话号码吗？`,
				i18n.LangIdID: `Apakah Anda tahu nomor telepon dari *** (Hubungi 2)? (Ini pertanyaan terbuka dan mengharuskan pemohon untuk memberikan nomor lengkap)`,
			},
			InputType: "radio",
			Field:     "contact2",
		},
	},

	ModuleOther: map[int]PhoneVerifyQuestionItem{
		14001: {
			QuestionSN: 14001,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您的放款银行卡是什么银行的？`,
				i18n.LangIdID: `Apa bank kartu bank Anda?`,
			},
			InputType: "radio",
			Field:     "bank_name",
		},
		14002: {
			QuestionSN: 14002,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您的放款银行卡卡号是？`,
				i18n.LangIdID: `Berapa nomor kartu bank Anda?`,
			},
			InputType: "radio",
			Field:     "bank_no",
		},
		14003: {
			QuestionSN: 14003,
			Question: i18n.LangItem{
				i18n.LangZhCN: `银行卡是您本人名下的吗？`,
				i18n.LangIdID: `Apakah kartu bank nama pribadi Anda?`,
			},
			InputType: "text",
			Field:     "",
		},
		14004: {
			QuestionSN: 14004,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您已经结婚了吗？`,
				i18n.LangIdID: `Apakah kamu sudah menikah?`,
			},
			InputType: "radio",
			Field:     "marital_status",
		},
		14005: {
			QuestionSN: 14005,
			Question: i18n.LangItem{
				i18n.LangZhCN: `如借款人未婚，则跳过此问题。您有几个孩子？`,
				i18n.LangIdID: `"Jika peminjam tidak menikah, lewati pertanyaan ini. Berapa banyak anak yang Anda miliki? """`,
			},
			InputType: "radio",
			Field:     "children_number",
		},
		14006: {
			QuestionSN: 14006,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您现在居住在哪里？是您自己的房子吗？`,
				i18n.LangIdID: `Di mana kamu tinggal sekarang?`,
			},
			InputType: "radio",
			Field:     "resident_address",
		},
		14007: {
			QuestionSN: 14007,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您的住处附近是不是有一个***（比如公园）？`,
				i18n.LangIdID: `Apakah ada **** (seperti taman) di dekat tempat tinggal Anda? (Untuk referensi ke bangunan landmark di sekitar alamat rumah pelanggan)`,
			},
			InputType: "text",
			Field:     "",
		},
		14008: {
			QuestionSN: 14008,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您是大学毕业吗？`,
				i18n.LangIdID: `Apakah kamu lulus dari perguruan tinggi?`,
			},
			InputType: "radio",
			Field:     "education",
		},
	},

	ModuleReloan: map[int]PhoneVerifyQuestionItem{
		16001: {
			QuestionSN: 16001,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您上次是使用什么品牌的手机进行的借款？`,
				i18n.LangIdID: `merek ponsel apa yang Anda pakai untuk pinjaman terakhir kali?`,
			},
			InputType: "radio",
			Field:     "last_time_cellphone_model",
		},

		16002: {
			QuestionSN: 16002,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您上次借款的金额是？`,
				i18n.LangIdID: `Berapa Jumlah yang Anda pinjam terakhir kali? `,
			},
			InputType: "radio",
			Field:     "last_time_loan_amount",
		},

		16003: {
			QuestionSN: 16003,
			Question: i18n.LangItem{
				i18n.LangZhCN: `您上次的借款用途是？`,
				i18n.LangIdID: `penggunaan pinjaman terakhir kali untuk apa ?`,
			},
			InputType: "text",
			Field:     "",
		},
	},
	ModuleInfoReview: map[int]PhoneVerifyQuestionItem{
		18001: {
			QuestionSN: 18001,
			Question: i18n.LangItem{
				i18n.LangZhCN: `身份证照片是否正常？`,
				i18n.LangIdID: `Apa Masalah pada Foto KTP?`,
			},
			InputType: "radio",
			Field:     "",
			Reasons: []i18n.LangItem{
				{
					ReasonSq:      "1",
					i18n.LangZhCN: `身份证破损`,
					i18n.LangIdID: `KTP rusak`,
				},
				{
					ReasonSq:      "2",
					i18n.LangZhCN: `身份证过期`,
					i18n.LangIdID: `KTP sudah kadaluwarsa`,
				},
				{
					ReasonSq:      "3",
					i18n.LangZhCN: `身份证照片伪造`,
					i18n.LangIdID: `Foto KTP palsu`,
				},
				{
					ReasonSq:      "4",
					i18n.LangZhCN: `身份证照片模糊`,
					i18n.LangIdID: `Foto KTP buram/kabur`,
				},
				{
					ReasonSq:      "5",
					i18n.LangZhCN: `其他`,
					i18n.LangIdID: `Lainnya`,
				},
			},
		},
		18002: {
			QuestionSN: 18002,
			Question: i18n.LangItem{
				i18n.LangZhCN: `手持证件照是否正常？`,
				i18n.LangIdID: `Apa Masalah pada Foto KTP yang Dipegang?`,
			},
			InputType: "radio",
			Field:     "",
			Reasons: []i18n.LangItem{
				{
					ReasonSq:      "1",
					i18n.LangZhCN: `手持证件照模糊`,
					i18n.LangIdID: `Foto KTP yang dipegang kabur/buram`,
				},
				{
					ReasonSq:      "2",
					i18n.LangZhCN: `缺少身份证`,
					i18n.LangIdID: `Bukan KTP`,
				},
				{
					ReasonSq:      "3",
					i18n.LangZhCN: `只有身份证`,
					i18n.LangIdID: `Hanya ada KTP`,
				},
				{
					ReasonSq:      "4",
					i18n.LangZhCN: `身份证照片模糊`,
					i18n.LangIdID: `Foto KTP buram/kabur`,
				},
				{
					ReasonSq:      "5",
					i18n.LangZhCN: `手持人与证件照不匹配`,
					i18n.LangIdID: `Foto di KTP dengan orang yang memegang tidak cocok`,
				},
				{
					ReasonSq:      "6",
					i18n.LangZhCN: `其他`,
					i18n.LangIdID: `Lainnya`,
				},
			},
		},
		18003: {
			QuestionSN: 18003,
			Question: i18n.LangItem{
				i18n.LangZhCN: `活体是否正常？`,
				i18n.LangIdID: `Apa Masalah pada Verifikasi Langsung (Live Verification)?`,
			},
			InputType: "radio",
			Field:     "",
			Reasons: []i18n.LangItem{
				{
					ReasonSq:      "1",
					i18n.LangZhCN: `活体照片与身份证不匹配`,
					i18n.LangIdID: `Foto verifikasi langsung tidak sesuai dengan KTP`,
				},
				{
					ReasonSq:      "2",
					i18n.LangZhCN: `其他`,
					i18n.LangIdID: `Lainnya`,
				},
			},
		},
		18004: {
			QuestionSN: 18004,
			Question: i18n.LangItem{
				i18n.LangZhCN: `联系人信息是否正常？`,
				i18n.LangIdID: `Apa Masalah pada Informasi Kontak yang Dapat Dihubungi?`,
			},
			InputType: "radio",
			Field:     "",
			Reasons: []i18n.LangItem{
				{
					ReasonSq:      "1",
					i18n.LangZhCN: `联系人缺少姓名`,
					i18n.LangIdID: `Tidak ada nama pada kontak yang dapat dihubungi`,
				},
				{
					ReasonSq:      "2",
					i18n.LangZhCN: `联系人姓名带有特殊符号`,
					i18n.LangIdID: `Terdapat simbol pada nama kontak yang dapat dihubungi`,
				},
				{
					ReasonSq:      "3",
					i18n.LangZhCN: `联系人电话异常`,
					i18n.LangIdID: `Nomor HP kontak yang dapat dihubungi tidak terdaftar`,
				},
				{
					ReasonSq:      "4",
					i18n.LangZhCN: `缺少联系人`,
					i18n.LangIdID: `Tidak ada kontak yang dapat dihubungi`,
				},
				{
					ReasonSq:      "5",
					i18n.LangZhCN: `其他`,
					i18n.LangIdID: `Lainnya`,
				},
			},
		},
		18005: {
			QuestionSN: 18005,
			Question: i18n.LangItem{
				i18n.LangZhCN: `本人电话是否注册为whatsapp账号？`,
				i18n.LangIdID: `apakah nomer hp terdaftar untuk whatsapp?`,
			},
			InputType: "radio",
			Field:     "",
			Reasons: []i18n.LangItem{
				{
					ReasonSq:      "1",
					i18n.LangZhCN: `是`,
					i18n.LangIdID: `Ya`,
				},
				{
					ReasonSq:      "2",
					i18n.LangZhCN: `否`,
					i18n.LangIdID: `Tidak`,
				},
			},
		},
	},
}

var fixedVerifyQuestionItems = []PhoneVerifyQuestionItem{
	{
		QuestionSN: 15001,
		Question: i18n.LangItem{
			i18n.LangZhCN: `电核电话分时段拨打是否正常接听?`,
			i18n.LangIdID: `Panggilan waktu normal dan normal?`,
		},
		InputType: "radio",
		Field:     "answer_phone_status",
	},
	{
		QuestionSN: 15002,
		Question: i18n.LangItem{
			i18n.LangZhCN: `身份信息是否正常（活体认证与身份证非同一人/证件照作假/未上传手持证件照）?`,
			i18n.LangIdID: `Informasi KTP normal apa tidak (foto dgn KTP tidak sesuai / KTP tidak benar / belum upload foto dengan KTP/KTP Tidak Berlaku)?`,
		},
		InputType: "radio",
		Field:     "identity_info_status",
	},
	{
		QuestionSN: 15003,
		Question: i18n.LangItem{
			i18n.LangZhCN: `您的手机号码是？`,
			i18n.LangIdID: `Silahkan katakan nomor ponsel Anda?`,
		},
		InputType: "radio",
		Field:     "owner_mobile_status",
	},
	/*
		{
			QuestionSN: 15004,
			Question: i18n.LangItem{
				i18n.LangZhCN: `本人电话是否注册为whatsapp账号？`,
				i18n.LangIdID: `apakah nomer hp terdaftar untuk whatsapp?`,
			},
			InputType: "radio",
			Field:     "owner_mobile_whatsapp",
		},
	*/
}

var fixedReloanVerifyQuestionItems = []PhoneVerifyQuestionItem{
	{
		QuestionSN: 17001,
		Question: i18n.LangItem{
			i18n.LangZhCN: `电核电话分时段拨打是否正常接听?`,
			i18n.LangIdID: `Panggilan waktu normal dan normal?`,
		},
		InputType: "radio",
		Field:     "answer_phone_status",
	},
	{
		QuestionSN: 17002,
		Question: i18n.LangItem{
			i18n.LangZhCN: ` 身份信息是否正常（复贷手持身份证照片和首贷是否一致/复贷手持身份证照片和活体是否一致）？`,
			i18n.LangIdID: `Informasi KTP normal apa tidak (foto dgn KTP tidak sesuai / KTP tidak benar / belum upload foto dengan KTP/KTP Tidak Berlaku)?`,
		},
		InputType: "radio",
		Field:     "identity_info_status",
	},
}

// 随机选3个模块
func choiceRandomModule() (m []ModuleSN) {
	//var i = 0
	for _, sn := range moduleSNMap {
		//if i >= 3 {
		//	break
		//}

		m = append(m, sn)

		//i++
	}

	return
}

// 复贷问题
func reloanChoiceModule() (m []ModuleSN) {
	for _, sn := range reloanModuleSNMap {
		m = append(m, sn)
	}

	return
}

// 通过qid取对应的问题
func GetPhoneVerifyQuestionById(lang, strID string) (question string, err error) {
	qid, _ := tools.Str2Int(strID)
	midInt, _ := tools.Str2Int(tools.SubString(strID, 0, 2))
	mid := ModuleSN(midInt)

	if strID == "0" {
		return
	}

	if _, ok := phoneVerifyQuestionConfig[mid]; !ok {
		err = fmt.Errorf("undefined module, qid: %s", strID)
		return
	}

	questionBox := phoneVerifyQuestionConfig[mid]
	if _, ok := questionBox[qid]; !ok {
		err = fmt.Errorf("undefined item, qid: %s", strID)
		return
	}

	questionItem := questionBox[qid]
	if _, ok := questionItem.Question[lang]; ok {
		question = questionItem.Question[lang]
	} else {
		question = questionItem.Question[i18n.LangZhCN]
	}

	return
}

func GetFixedPhoneVerifyQuestionByOffset(lang string, offset int) (question string, err error) {
	if offset < 0 || offset >= len(fixedVerifyQuestionItems) {
		err = fmt.Errorf("out of range, offset: %d", offset)
		return
	}

	questionItem := fixedVerifyQuestionItems[offset]
	if _, ok := questionItem.Question[lang]; ok {
		question = questionItem.Question[lang]
	} else {
		question = questionItem.Question[i18n.LangZhCN]
	}

	return
}

func PhoneVerifyQuestionItemTrans(lang string, item i18n.LangItem) (q string) {
	if v, ok := item[lang]; ok {
		q = v
	} else if v, ok := item[i18n.LangZhCN]; ok {
		q = v
	} else {
		logs.Error("[PhoneVerifyQuestionItemTrans] has wrong, lang:", lang, ", item:", item)
	}

	return
}

type RiskCtlQuerySet struct {
	OrderId           int64
	AccountId         int64
	Realname          string
	Amount            int64 // 借款金额
	Loan              int64 // 放贷金额
	IsReloan          types.IsReloanEnum
	Period            int // 借款期限
	CheckStatus       types.LoanStatus
	RiskCtlStatus     types.RiskCtlEnum
	RiskCtlFinishTime int64
	PhoneVerifyTime   int64
	RiskCtlRegular    string
	CheckTime         int64
	ApplyTime         int64
	RandomValue       int
	FixedRandom       int
	OpUid             int64
	LoanTime          int64
	FinishTime        int64
	PlatformMark      int64
}

func RiskCtlList(condCntr map[string]interface{}, page, pagesize int) (count int64, list []RiskCtlQuerySet, num int64, err error) {
	// 需要join多表: orders o, account_base a
	accountBase := models.AccountBase{}
	orders := models.Order{}
	o := orm.NewOrm()
	o.Using(orders.UsingSlave())

	var sql string
	// 风控只关心非临时,有风控状态的订单
	var where string = fmt.Sprintf("WHERE o.risk_ctl_status > 0 AND o.is_temporary = %d", types.IsTemporaryNO)

	sqlCount := "SELECT COUNT(o.id) AS total"
	sqlQuery := `SELECT o.id AS order_id, a.id AS account_id, a.realname, a.platform_mark, o.amount,
o.loan, o.is_reloan, o.loan_time,o.finish_time,o.check_time, o.period,
o.check_status, o.risk_ctl_status, o.risk_ctl_finish_time,
o.phone_verify_time, o.apply_time, o.op_uid, o.risk_ctl_regular,
o.random_value, o.fixed_random`
	from := fmt.Sprintf("FROM `%s` o LEFT JOIN `%s` a ON o.user_account_id = a.id",
		orders.TableName(), accountBase.TableName())

	if v, ok := condCntr["realname"]; ok {
		where = fmt.Sprintf("%s AND a.realname = '%s'", where, tools.Escape(v.(string)))
	}
	if v, ok := condCntr["risk_ctl_regular"]; ok {
		where = fmt.Sprintf("%s AND o.risk_ctl_regular = '%s'", where, tools.Escape(v.(string)))
	}
	if f, ok := condCntr["id"]; ok {
		where = fmt.Sprintf("%s AND o.id = %d", where, f.(int64))
	}
	if f, ok := condCntr["risk_ctl_status"]; ok {
		checkStatusArr := f.([]string)
		if len(checkStatusArr) > 0 {
			where = fmt.Sprintf("%s AND o.risk_ctl_status IN (%s)", where, strings.Join(checkStatusArr, ", "))
		}
	}
	if _, ok := condCntr["apply_time_start"]; ok {
		where = fmt.Sprintf("%s AND o.apply_time >= %d AND o.apply_time <= %d", where, condCntr["apply_time_start"].(int64), condCntr["apply_time_end"].(int64))
	}
	if _, ok := condCntr["check_time_start"]; ok {
		where = fmt.Sprintf("%s AND o.check_time >= %d AND o.check_time <= %d", where, condCntr["check_time_start"].(int64), condCntr["check_time_end"].(int64))
	}
	if f, ok := condCntr["user_account_id"]; ok {
		where = fmt.Sprintf("%s AND o.user_account_id = %d", where, f.(int64))
	}
	if f, ok := condCntr["is_reloan"]; ok {
		where = fmt.Sprintf("%s AND o.is_reloan = %d", where, f.(int64))
	}
	if f, ok := condCntr["random_value_start"]; ok {
		where = fmt.Sprintf("%s AND o.random_value >= %d", where, f.(int64))
	}
	if f, ok := condCntr["random_value_end"]; ok {
		where = fmt.Sprintf("%s AND o.random_value <= %d", where, f.(int64))
	}
	if f, ok := condCntr["fix_value_start"]; ok {
		where = fmt.Sprintf("%s AND o.fixed_random >= %d", where, f.(int64))
	}
	if f, ok := condCntr["fix_value_end"]; ok {
		where = fmt.Sprintf("%s AND o.fixed_random <= %d", where, f.(int64))
	}
	if f, ok := condCntr["platform_mark"]; ok {
		platformMark := f.([]string)
		mark := int64(0)
		for _, v := range platformMark {
			i, _ := tools.Str2Int64(v)
			mark = mark | i
		}

		if mark > 0 {
			where = fmt.Sprintf("%s AND a.platform_mark & %d > 0", where, mark)
		}
	}

	//// 金管局需求,有要求时打开
	//where = fmt.Sprintf("%s AND ( a.is_deleted = 0 and o.is_deleted = 0 )", where)

	orderBy := ""
	if v, ok := condCntr["field"]; ok {
		if vF, okF := fieldMap[v.(string)]; okF {
			orderBy = "ORDER BY " + vF
		} else {
			orderBy = "ORDER BY o.id"
		}
	} else {
		orderBy = "ORDER BY o.id"
	}

	if v, ok := condCntr["sort"]; ok {
		orderBy = fmt.Sprintf("%s %s", orderBy, v.(string))
	} else {
		orderBy = fmt.Sprintf("%s %s", orderBy, "DESC")
	}

	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = Pagesize
	}
	offset := (page - 1) * pagesize
	limit := fmt.Sprintf("LIMIT %d, %d", offset, pagesize)

	sql = fmt.Sprintf("%s %s %s", sqlCount, from, where)
	o.Raw(sql).QueryRow(&count)

	sql = fmt.Sprintf("%s %s %s %s %s", sqlQuery, from, where, orderBy, limit)
	num, err = o.Raw(sql).QueryRows(&list)

	return
}

// GetAllHitRegularByOrderID 取订单命中的所有风控策略,
// 且属于当前订单审核规则列表中的规则,即 status = types.RiskCtlRegularReviewed
func GetAllHitRegularByOrderID(orderID int64) (list []string, num int64, err error) {
	var listRegular []models.RiskRegularRecord

	oneRegular := models.RiskRegularRecord{}
	o := orm.NewOrm()
	o.Using(oneRegular.UsingSlave())

	num, err = o.QueryTable(oneRegular.TableName()).
		Filter("order_id", orderID).Filter("status", types.RiskCtlRegularReviewed).All(&listRegular)

	if num > 0 {
		for _, one := range listRegular {
			list = append(list, one.HitRegular)
		}
	}

	return
}

// 获得所有的记录
func GetAllHitRegularRecordByOrderID(orderID int64) (listRegular []models.RiskRegularRecord, num int64, err error) {
	oneRegular := models.RiskRegularRecord{}
	o := orm.NewOrm()
	o.Using(oneRegular.UsingSlave())

	num, err = o.QueryTable(oneRegular.TableName()).
		Filter("order_id", orderID).Filter("status", types.RiskCtlRegularReviewed).OrderBy("id").All(&listRegular)

	return
}

type PhoneVerifyResultDetailItem struct {
	Question string `json:"question"`
	Status   string `json:"status"`
	Value    string `json:"value"`
	Qid      string `json:"qid"`
}

func getPhoneVerifyResultStatus(lang, stat string) (status string) {
	statKey, _ := tools.Str2Int(stat)
	if v, ok := questionStatusMap[statKey]; ok {
		status = v
	}

	return i18n.T(lang, status)
}

// 取出关组装一条电核结果
func GetPhoneVerifyResultDetail(lang string, orderID int64) (list []PhoneVerifyResultDetailItem, invalidReason string, remark string, err error) {
	r := models.PhoneVerifyRecord{}
	o := orm.NewOrm()
	o.Using(r.UsingSlave())

	var dbResult []orm.Params
	sql := fmt.Sprintf(`SELECT * FROM %s WHERE order_id = %d ORDER BY id DESC LIMIT 1`, r.TableName(), orderID)
	num, err := o.Raw(sql).Values(&dbResult)
	if err != nil || num <= 0 {
		logs.Warning("num:", num, ", err:", err)
		err = fmt.Errorf("no data")
		return
	}

	//logs.Debug("dbResult: %#v", dbResult[0])

	dbResultKV := dbResult[0]
	var item PhoneVerifyResultDetailItem

	// 固定问题1
	item.Question, _ = GetFixedPhoneVerifyQuestionByOffset(lang, 0)
	item.Status = getPhoneVerifyResultStatus(lang, dbResultKV["answer_phone_status"].(string))
	item.Value = ""
	list = append(list, item)

	// 固定问题2
	item.Question, _ = GetFixedPhoneVerifyQuestionByOffset(lang, 1)
	item.Status = getPhoneVerifyResultStatus(lang, dbResultKV["identity_info_status"].(string))
	item.Value = ""
	list = append(list, item)

	// 固定问题3
	item.Question, _ = GetFixedPhoneVerifyQuestionByOffset(lang, 2)
	item.Status = getPhoneVerifyResultStatus(lang, dbResultKV["owner_mobile_status"].(string))
	item.Value = ""
	list = append(list, item)

	/*
		// 固定问题4
		item.Question, _ = GetFixedPhoneVerifyQuestionByOffset(lang, 3)
		item.Status = getPhoneVerifyResultStatus(lang, dbResultKV["owner_mobile_whatsapp"].(string))
		item.Value = ""
		list = append(list, item)
	*/

	for i := 1; i <= 6; i++ {
		qidField := fmt.Sprintf("q%d_id", i)
		qStatusField := fmt.Sprintf("q%d_status", i)
		qValueField := fmt.Sprintf("q%d_value", i)

		//logs.Debug("qidField:", qidField, ", qStatusField:", qStatusField, "qValueField:", qValueField)

		item.Qid = dbResultKV[qidField].(string)
		item.Question, err = GetPhoneVerifyQuestionById(lang, dbResultKV[qidField].(string))
		//logs.Debug("err:", err)
		item.Status = getPhoneVerifyResultStatus(lang, dbResultKV[qStatusField].(string))
		if dbResultKV[qValueField] != nil {
			item.Value = dbResultKV[qValueField].(string)
		} else {
			item.Value = ""
		}
		list = append(list, item)
	}

	if dbResultKV["remark"] != nil {
		remark = dbResultKV["remark"].(string)
	}

	// 固定问题2
	invalidReason = dbResultKV["invalid_reason"].(string)

	return
}

// 风控策略---按级别

const (
	//FixedRiskCtlRegularRandomDefault int = 99
	//FixedPhoneVerifyRandomDefault    int = 999
	// FixedDayLimitOrdersReject 日订单熔断拒绝
	FixedDayLimitOrdersReject int = 111
	// FixedRiskCtlRegularRandom1 命中一类反欺诈规则,随机值失效
	FixedRiskCtlRegularRandom1 int = 8888
	// FixedRiskCtlRegularRandom2 命中二类反欺诈规则,随机值失效
	FixedRiskCtlRegularRandom2 int = 9999
	// FixedPhoneVerifyRandom 命中一级随机数, 但固定问题未通过 || 二级随机数但命中二级电核问题
	FixedPhoneVerifyRandom int = 99999
	// FixedPhoneVerifySet2Invalid 设置为无效订单
	FixedPhoneVerifySet2Invalid int = 444
)

// 电核问题级别1级
var phoneVerifyLevel1 = map[int]bool{
	15001: true,
}

func PhoneVerifyLevel1() map[int]bool {
	return phoneVerifyLevel1
}

//// 电核问题级别2级
func PhoneVerifyLevel2() map[int]bool {
	var phoneVerifyLevel2 = map[int]bool{}
	phoneVerifyLevel2Conf := strings.Split(config.ValidItemString("phone_verify_level_2"), ",")
	for _, qidStr := range phoneVerifyLevel2Conf {
		qid, _ := tools.Str2Int(qidStr)

		phoneVerifyLevel2[qid] = true
	}

	return phoneVerifyLevel2
}

// InfoReview自动外呼的A卡分数区间判断是否满足'等待自动外呼'
// '-' 或 '-,-' 都按没有配置处理，走默认流程
// '-,610' 按照没有下限处理
// '610,-' 按照没有上限处理
func isSatisfyAScoreConfig(orderData models.Order) (isSatisfy bool) {
	aScoreConfig := config.ValidItemString("inforeview_a_score_range")
	if aScoreConfig == "-" {
		return
	}

	var lower int
	var upper int
	scores := strings.Split(aScoreConfig, ",")
	if len(scores) == 1 {
		return
	}

	if scores[0] == "-" {
		lower = types.RiskCtlAOrBScoreLower // A/B卡分数在0-1500之内
	} else {
		lower, _ = strconv.Atoi(scores[0])
	}

	if scores[1] == "-" {
		upper = types.RiskCtlAOrBScoreUpper
	} else {
		upper, _ = strconv.Atoi(scores[1])
	}

	if lower == types.RiskCtlAOrBScoreLower && upper == types.RiskCtlAOrBScoreUpper {
		return
	}

	score := ticket.CalculateRiskScore(orderData.Id)
	logs.Info("[isSatisfyAScoreConfig] Risk score:", score, ", orderId:", orderData.Id)

	if (lower == types.RiskCtlAOrBScoreLower && lower < score && upper >= score) ||
		(upper == types.RiskCtlAOrBScoreUpper && score < upper && lower <= score) ||
		(lower <= score && upper >= score) {
		isSatisfy = true
	}
	logs.Info("[isSatisfyAScoreConfig] isSatisty:", isSatisfy)

	return
}

// PhoneVerifySave 保存电核结果, 并更新订单
func PhoneVerifySave(phoneVerify models.PhoneVerifyRecord, orderData models.Order, ticketInfo models.Ticket,
	redirectReject int, qid2StatusMap map[int]int, adminUID int64) error {
	models.AddOnePhoneVerifyRecord(phoneVerify)
	originOrder := orderData

	// 处理订单状态
	//// 审批不通过
	if redirectReject == 1 || phoneVerify.Result == 2 {
		orderData.CheckStatus = types.LoanStatusReject           // 审核拒绝
		orderData.RiskCtlStatus = types.RiskCtlPhoneVerifyReject // 电核拒绝
		orderData.RejectReason = types.RejectReasonVerifyFail    // 电核拒绝的通用原因
	} else if phoneVerify.Result == 1 { //// 审批通过
		orderData.CheckStatus = types.LoanStatusWait4Loan
		orderData.RiskCtlStatus = types.RiskCtlPhoneVerifyPass

		if ticketInfo.ItemID == types.TicketItemInfoReview && isSatisfyAScoreConfig(orderData) { // InfoReview审核
			orderData.CheckStatus = types.LoanStatusWaitAutoCall
			orderData.RiskCtlStatus = types.RiskCtlWaitAutoCall
		}
	} else if phoneVerify.Result == 3 { //// 重置为无效订单
		orderData.CheckStatus = types.LoanStatusInvalid          // 无效订单
		orderData.RiskCtlStatus = types.RiskCtlPhoneVerifyReject // 电核拒绝
		orderData.RejectReason = types.RejectReasonVerifyFail    // 电核拒绝的通用原因
		orderData.FixedRandom = FixedPhoneVerifySet2Invalid      // 消息随机策略

		// 清空客户资料,使其重新上传证件照和手持照
		accountBase, _ := models.OneAccountBaseByPkId(orderData.UserAccountId)
		originData := accountBase

		accountBase.Realname = ""
		accountBase.Identity = ""
		accountBase.Update("realname", "identity")

		// 操作流水
		models.OpLogWrite(adminUID, orderData.UserAccountId, models.OpCodeAccountBaseUpdate, accountBase.TableName(), originData, accountBase)

		accountCoupon, err := dao.GetAccountFrozenCouponByOrder(orderData.UserAccountId, orderData.Id)
		if err == nil {
			MakeAccountCouponAvailable(&accountCoupon)
		}

		AddAccountBaseExtPhyInvalidTag(orderData.UserAccountId, adminUID)
		AddOrdersExtPhyInvalidTag(orderData.Id, adminUID)

	}

	//如果因电话异常拒绝，打客户召回标签
	if phoneVerify.AnswerPhoneStatus == 2 {

		log, _ := models.OneRecallPhoneVerifyTagLogByAOID(orderData.UserAccountId, orderData.Id)
		log.Remark = 1
		log.Utime = tools.GetUnixMillis()
		cols := []string{"remark", "utime"}
		models.OrmUpdate(&log, cols)
		logs.Debug("[PhoneVerifySave]因电话异常拒绝时，打remark标记")

		timetag := tools.NaturalDay(0)
		CustomerRecallTag(timetag)
	}

	if orderData.CheckStatus != types.LoanStatusWaitAutoCall && orderData.RiskCtlStatus != types.RiskCtlWaitAutoCall {
		// 随机数的分级策略
		if orderData.FixedRandom <= 0 {
			if IsLevel1Random(orderData.RandomValue) {
				// 如果一级随机数命中且有效,并且电话有人接听,则可以忽略电核结论,给放款
				if phoneVerify.AnswerPhoneStatus == QuestionStatusNormal &&
					phoneVerify.IdentityInfoStatus == QuestionStatusNormal &&
					phoneVerify.OwnerMobileStatus == QuestionStatusNormal {
					orderData.CheckStatus = types.LoanStatusWait4Loan
				} else {
					// 电话无人接听
					orderData.CheckStatus = types.LoanStatusReject
					orderData.FixedRandom = FixedPhoneVerifyRandom
				}
			} else if IsLevel2Random(orderData.RandomValue) { // 命中二级随机数
				// 过一遍二级电核规则
				var isHit2 = false
				var phoneVerifyLevel2 = PhoneVerifyLevel2()
				for questionID, qStatus := range qid2StatusMap {
					if qStatus == QuestionStatusAbnormal {
						if phoneVerifyLevel2[questionID] {
							isHit2 = true
							break
						}
					}
				}

				if phoneVerify.AnswerPhoneStatus == QuestionStatusNormal &&
					phoneVerify.IdentityInfoStatus == QuestionStatusNormal &&
					phoneVerify.OwnerMobileStatus == QuestionStatusNormal {
					if isHit2 { // 命中二级电核
						orderData.CheckStatus = types.LoanStatusReject
						orderData.FixedRandom = FixedPhoneVerifyRandom
					} else {
						// 过了所有二电核
						orderData.CheckStatus = types.LoanStatusWait4Loan
					}
				} else {
					// 电话无人接听
					orderData.CheckStatus = types.LoanStatusReject
					orderData.FixedRandom = FixedPhoneVerifyRandom
				}
			}
		}
	}

	if orderData.CheckStatus == types.LoanStatusWait4Loan && phoneVerify.Result != 1 {
		markCustomerAsRandomHit(orderData.UserAccountId, orderData.Id, adminUID)
	}

	if orderData.CheckStatus == types.LoanStatusWait4Loan {
		CheckThirdBlacklist(&orderData)
	}

	if orderData.CheckStatus == types.LoanStatusWait4Loan {
		//第三方黑名单通过后增加人脸比对
		CompareAfterBlackList(&orderData)
	}

	if orderData.CheckStatus == types.LoanStatusWait4Loan {
		schema_task.PushBusinessMsg(types.PushTargetReviewPass, orderData.UserAccountId)
	} else if orderData.CheckStatus == types.LoanStatusReject {
		schema_task.PushBusinessMsg(types.PushTargetReviewReject, orderData.UserAccountId)
	} else if orderData.CheckStatus == types.LoanStatusInvalid {
		switch phoneVerify.InvalidReason {
		case types.InvalidIdentifyFog:
			schema_task.PushBusinessMsg(types.PushTargetInvalidFog, orderData.UserAccountId)
		case types.InvalidIdentifyHoldFog:
			schema_task.PushBusinessMsg(types.PushTargetInvalidHoldFog, orderData.UserAccountId)
		case types.InvalidIdentifyHoldNoFace:
			schema_task.PushBusinessMsg(types.PushTargetInvalidHoldNoFace, orderData.UserAccountId)
		case types.InvalidIdentifyHoldNoIdentify:
			schema_task.PushBusinessMsg(types.PushTargetInvalidHoldNoIdentify, orderData.UserAccountId)
		}
	}

	monitor.IncrOrderCount(orderData.CheckStatus)

	//// 订单审核需要修改的其他字段
	orderData.CheckTime = tools.GetUnixMillis()
	orderData.OpUid = adminUID
	orderData.PhoneVerifyTime = tools.GetUnixMillis()
	orderData.Utime = tools.GetUnixMillis()

	_, err := models.UpdateOrder(&orderData)
	if err != nil {
		return err
	}

	//完成工单
	//ticket.CompleteByRelatedID(orderData.Id, types.TicketItemPhoneVerify)
	ticket.CompletePhoneVerifyOrInfoReviewByRelatedID(orderData.Id)

	// 写修改订单数据的操作日志
	models.OpLogWrite(adminUID, orderData.Id, models.OpCodeOrderUpdate, orderData.TableName(), originOrder, orderData)
	return nil
}

func AddAccountBaseExtPhyInvalidTag(accountId int64, adminUID int64) {
	aExt, _ := models.OneAccountBaseExtByPkId(accountId)
	if aExt.PhyInvalidTag == types.PhoneVerifyInvalidTag {
		logs.Info("[AddAccountBaseExtPhyInvalidTag] no need to add tag. accountId:%d", accountId)
		return
	}

	old := aExt
	tag := tools.GetUnixMillis()
	aExt.PhyInvalidTag = types.PhoneVerifyInvalidTag
	aExt.Utime = tag

	if aExt.AccountId == 0 {
		aExt.AccountId = accountId
		aExt.Ctime = tag
		models.OrmInsert(&aExt)
	} else {
		cols := []string{"utime", "phy_invalid_tag"}
		models.OrmUpdate(&aExt, cols)
	}
	models.OpLogWrite(adminUID, accountId, models.OpCodeOrderUpdate, aExt.TableName(), old, aExt)
}

func tryAddOrderPhyInvalidTag(accountId int64, orderId int64) {
	aExt, _ := models.OneAccountBaseExtByPkId(accountId)
	reLoan := dao.IsRepeatLoan(accountId)
	if aExt.PhyInvalidTag != types.PhoneVerifyInvalidTag || reLoan {
		logs.Info("[tryAddOrderPhyInvalidTag] accountId:%d no phy tag or reLoan:%v. order:%d no need to add tag.", accountId, reLoan, orderId)
		return
	}
	logs.Info("[tryAddOrderPhyInvalidTag] accountId:%d add  phy tag. order:%d .", accountId, orderId)
	AddOrdersExtPhyInvalidTag(orderId, 0)
}

func AddOrdersExtPhyInvalidTag(orderId int64, adminUID int64) {
	oExt, _ := models.GetOrderExt(orderId)
	if oExt.PhyInvalidTag == types.PhoneVerifyInvalidTag {
		logs.Info("[AddOrdersExtPhyInvalidTag] no need to add tag. orderId:%d", orderId)
		return
	}

	old := oExt
	tag := tools.GetUnixMillis()
	oExt.PhyInvalidTag = types.PhoneVerifyInvalidTag
	oExt.Utime = tag

	if oExt.OrderId == 0 {
		oExt.OrderId = orderId
		oExt.Ctime = tag
		models.OrmInsert(&oExt)
	} else {
		cols := []string{"utime", "phy_invalid_tag"}
		models.OrmUpdate(&oExt, cols)
	}
	models.OpLogWrite(adminUID, orderId, models.OpCodeOrderUpdate, oExt.TableName(), old, oExt)
}

// 反欺诈规则一级规则
func RiskCtlRegularLevel1() (riskCtlRegularLevel1 map[string]bool, err error) {
	riskCtlRegularLevel1 = make(map[string]bool)

	originConf := config.ValidItemString("risk_ctl_regular_level_1")
	if len(originConf) <= 0 {
		err = fmt.Errorf("[RiskCtlRegularLevel1] can not find [risk_ctl_regular_level_1] config from system")
		logs.Warning("err:", err)
		return
	}

	riskCtlRegularLevel1Conf := strings.Split(originConf, ",")
	for _, regular := range riskCtlRegularLevel1Conf {
		riskCtlRegularLevel1[regular] = true
	}

	return
}

//// 反欺诈规则,以下规则为2级
func RiskCtlRegularLevel2() (riskCtlRegularLevel2 map[string]bool, err error) {
	riskCtlRegularLevel2 = make(map[string]bool)

	originConf := config.ValidItemString("risk_ctl_regular_level_2")
	if len(originConf) <= 0 {
		err = fmt.Errorf("[RiskCtlRegularLevel2] can not find [risk_ctl_regular_level_2] config from system")
		logs.Warning("err:", err)
		return
	}

	riskCtlRegularLevel2Conf := strings.Split(originConf, ",")
	for _, regular := range riskCtlRegularLevel2Conf {
		riskCtlRegularLevel2[regular] = true
	}

	return
}

//
func GetUnservicedAreaConf() (conf map[string]bool, err error) {
	conf = make(map[string]bool)

	originConf := config.ValidItemString("unserviced_area_conf")
	if len(originConf) <= 0 {
		err = fmt.Errorf("[GetUnservicedArearConf] can not find [unserviced_area_conf] config from system")
		logs.Warning("err:", err)
		return
	}

	confBox := strings.Split(originConf, ",")
	for _, v := range confBox {
		conf[v] = true
	}

	return
}

// 2018-04-20 随机数比例一级调整为0%，二级调整为1% {{{
//[1,100]
// 是否为一级随机数
func IsLevel1Random(randomValue int) (yes bool) {
	randomRate, _ := config.ValidItemInt("random_level_1")
	yes = false
	if randomValue > 100-randomRate {
		yes = true
	}

	return
}

// 是否为二级随机数
func IsLevel2Random(randomValue int) (yes bool) {
	randomRate, _ := config.ValidItemInt("random_level_2")
	yes = false
	if randomValue < 1+randomRate {
		yes = true
	}

	return
}

// }}}

// GetTodayLoanOrderTotal 取当日的放款订单总数,不要求十分精确
func GetTodayLoanOrderTotal() (num int64, err error) {
	startTime := tools.NaturalDay(0) + 3600*1000

	order := models.Order{}
	o := orm.NewOrm()
	o.Using(order.UsingSlave())

	//sql := fmt.Sprintf("SELECT COUNT(id) AS total FROM %s WHERE check_status = %d AND loan_time >= %d",
	sql := fmt.Sprintf("SELECT COUNT(id) AS total FROM %s WHERE loan_time >= %d",
		order.TableName(), startTime)
	err = o.Raw(sql).QueryRow(&num)

	return
}

func MarkCustomerIfHitRandom(orderData models.Order) (success bool, err error) {
	orderDataJSON, _ := tools.JsonEncode(orderData)
	if orderData.Id <= 0 || orderData.LoanTime <= 0 {
		logs.Error("[MarkCustomerIfHitRandom] 订单数据有误, orderData:", orderDataJSON)
		return
	}

	if len(orderData.RiskCtlRegular) > 0 {
		// 说明是随机数放款,那么,标记一下
		markCustomerAsRandomHit(orderData.UserAccountId, orderData.Id, 0)
		// accountBase, _ := models.OneAccountBaseByPkId(orderData.UserAccountId)
		// origin := accountBase
		//
		// accountBase.RandomMark = orderData.Id
		// accountBase.Update("random_mark")
		//
		// models.OpLogWrite(0, models.OpCodeAccountBaseUpdate, accountBase.TableName(), origin, accountBase)
	}

	success = true

	return
}

//
func markCustomerAsRandomHit(uid, orderID, adminUID int64) error {
	accountBase, _ := models.OneAccountBaseByPkId(uid)
	origin := accountBase

	accountBase.RandomMark = orderID
	result, err := accountBase.Update("RandomMark")
	models.OpLogWrite(0, accountBase.Id, models.OpCodeAccountBaseUpdate, accountBase.TableName(), origin, accountBase)
	if err != nil {
		logs.Error("MarkCustomerAsRandomHit update error:", err)
		return err
	}
	if result != 1 {
		err = errors.New("no sql error , but no rows affected")
		logs.Error("MarkCustomerAsRandomHit update error:", err)
		return err
	}

	return nil
}

func IsSkipPhoneVerify(randomValue int) bool {
	randomRate, _ := config.ValidItemInt("random_skip_phone_verify")
	yes := false
	randomValue = randomValue - 50
	if randomValue > 0 && randomValue < 1+randomRate {
		yes = true
	}

	logs.Debug("[IsSkipPhoneVerify] random:%d, rate:%d ret:%t", randomValue, randomRate, yes)

	return yes
}

// CompareAfterBlackList 订单流程第三方黑名单后进行人脸比对
func CompareAfterBlackList(orderData *models.Order) {
	logs.Debug("[CompareAfterBlackList] accountId:%d, orderId:%d, orderStatus:%d", orderData.UserAccountId, orderData.Id, orderData.CheckStatus)

	if orderData.CheckStatus != types.LoanStatusWait4Loan && orderData.CheckStatus != types.LoanStatusWaitPhotoCompare {
		return
	}
	//比对开关
	compareSwitch, _ := config.ValidItemBool("compare_after_blacklist_switch")
	if !compareSwitch {
		logs.Debug("[CompareAfterBlackList] compare switch turn off ,skip photo compare  accountID:%d,orderID:%d", orderData.UserAccountId, orderData.Id)
		return
	}
	if orderData.CheckStatus != types.LoanStatusWaitPhotoCompare {
		var originOrder models.Order = *orderData
		orderData.CheckStatus = types.LoanStatusWaitPhotoCompare
		orderData.RiskCtlStatus = types.RiskCtlWaitPhotoCompare
		orderData.Utime = tools.GetUnixMillis()
		models.UpdateOrder(orderData)
		monitor.IncrOrderCount(orderData.CheckStatus)
		// 添加操作日志
		models.OpLogWrite(0, orderData.Id, models.OpCodeOrderUpdate, orderData.TableName(), originOrder, *orderData)
	}

	similarity := CompareIDPhotoAndLivingEnv(orderData.UserAccountId)
	configSimilarity, _ := config.ValidItemFloat64("compare_liveenv_idphoto_val")

	logs.Debug("[CompareAfterBlackList] accountID:%d,orderID:%d, similarity:%g, configSimilarity:%g", orderData.UserAccountId, orderData.Id, similarity, configSimilarity)
	if similarity > configSimilarity {
		orderData.CheckStatus = types.LoanStatusWait4Loan
		orderData.RiskCtlStatus = types.RiskCtlPhotoComparePass
	} else {
		orderData.CheckStatus = types.LoanStatusReject
		orderData.RiskCtlStatus = types.RiskCtlPhotoCompareFail
		orderData.RejectReason = types.RejectReasonLackCredit
	}
	orderData.AfterBlackSimilar = tools.Float642Str(similarity)
}

func RecheckThirdBlacklist(uid int64, orderId int64) {
	logs.Info("[RecheckThirdBlacklist] orderId:%d", orderId)

	orderData, err := models.GetOrder(orderId)
	if err != nil {
		logs.Warn("[RecheckThirdBlacklist] GetOrder error orderId:%d, err:%v", orderId, err)
		return
	}

	if orderData.RiskCtlStatus != types.RiskCtlThirdBlacklistDoing {
		logs.Warn("[RecheckThirdBlacklist] Order status error orderId:%d, err:%d", orderId, orderData.CheckStatus)
		return
	}

	originOrder := orderData

	CheckThirdBlacklist(&orderData)
	if orderData.CheckStatus == types.LoanStatusWait4Loan {
		//第三方黑名单通过后增加人脸比对
		CompareAfterBlackList(&orderData)
	}
	orderData.CheckTime = tools.GetUnixMillis()
	orderData.OpUid = uid
	orderData.Utime = tools.GetUnixMillis()

	_, err = models.UpdateOrder(&orderData)
	if err != nil {
		return
	}

	// 写修改订单数据的操作日志
	models.OpLogWrite(uid, orderData.Id, models.OpCodeOrderUpdate, orderData.TableName(), originOrder, orderData)
	return
}

func CheckThirdBlacklist(orderData *models.Order) {
	logs.Debug("[CheckThirdBlacklist] accountId:%d, orderId:%d, orderStatus:%d", orderData.UserAccountId, orderData.Id, orderData.CheckStatus)

	if orderData.CheckStatus != types.LoanStatusWait4Loan && orderData.CheckStatus != types.LoanStatusThirdBlacklistIsDoing {
		return
	}

	if orderData.CheckStatus != types.LoanStatusThirdBlacklistIsDoing {
		var originOrder models.Order = *orderData

		orderData.CheckStatus = types.LoanStatusThirdBlacklistIsDoing
		orderData.RiskCtlStatus = types.RiskCtlThirdBlacklistDoing
		// 添加必更新字段

		orderData.Utime = orderData.CheckTime
		models.UpdateOrder(orderData)

		monitor.IncrOrderCount(orderData.CheckStatus)

		// 添加操作日志
		models.OpLogWrite(0, orderData.Id, models.OpCodeOrderUpdate, orderData.TableName(), originOrder, *orderData)
	}

	accountBase, _ := models.OneAccountBaseByPkId(orderData.UserAccountId)

	number := areacode.PhoneWithoutServiceRegionCode(accountBase.Mobile)
	region := areacode.GetRegionCode()
	region = "+" + region
	isPass, err := CheckAdvanceBlacklist(orderData.Id, orderData.UserAccountId, accountBase.Realname, accountBase.Identity, region, number)
	if err != nil {
		logs.Error("[CheckThirdBlacklist] checkAdvanceBlacklist return error orderId:%d, accountId:%d, err:%v",
			orderData.Id, orderData.UserAccountId, err)
	}

	logs.Info("[CheckThirdBlacklist] checkAdvanceBlacklist orderId:%d, accountId:%d, ispass:%t",
		orderData.Id, orderData.UserAccountId, isPass)

	if isPass {
		orderData.CheckStatus = types.LoanStatusWait4Loan
		orderData.RiskCtlStatus = types.RiskCtlThirdBlacklistPass
	} else {
		orderData.CheckStatus = types.LoanStatusReject
		orderData.RiskCtlStatus = types.RiskCtlThirdBlacklistReject
		orderData.RejectReason = types.RejectReasonLackCredit

		event.Trigger(&evtypes.BlacklistEv{accountBase.Id, types.RiskItemIdentity, accountBase.Identity, types.RiskReasonAdvance, "Advance"})
	}
}

func CheckAdvanceBlacklist(orderId, accountId int64, name, identity, countryCode, mobile string) (bool, error) {
	enable, _ := config.ValidItemInt("advance_blacklist_enable")
	if enable == 0 {
		return true, nil
	}

	_, resCodeMap, err := getAdvanceBlacklist(orderId, accountId, name, identity, countryCode, mobile)
	if err != nil {
		return true, err
	}

	// 不再使用recommendation返回的结果
	//riskResStr := config.ValidItemString("advance_blacklist_list")
	//riskList := strings.Split(riskResStr, ",")
	//for _, risk := range riskList {
	//	if cmd == strings.Trim(risk, " ") {
	//		return false, nil
	//	}
	//}

	// 使用reasoncode 判断
	resCodeStr := config.ValidItemString("advance_blacklist_list_reason_code")
	resCodeConfigList := strings.Split(resCodeStr, ",")
	for _, risk := range resCodeConfigList {
		resCode := strings.Trim(risk, " ")
		if _, ok := resCodeMap[resCode]; ok {
			// 返回的code 包含后台配置信息 则命中 不通过
			return false, nil
		}
	}
	return true, nil
}

func getAdvanceBlacklist(orderId, accountId int64, name, identity, countryCode, mobile string) (string, map[string]bool, error) {
	m, err := models.GetAdvanceBlacklist(accountId, orderId)
	if err == nil {
		expire := tools.NaturalDay(-30)
		if m.Ctime > expire {
			return advance.BalcklistPass(&m)
		}
	}

	body, _, err := advance.BlacklistCheck(accountId, name, identity, countryCode, mobile)
	str := string(body)

	logs.Debug("[CheckAdvanceBlacklist] BlacklistCheck return orderId:%d, accountId:%d, name:%s, identity:%s, countryCode:%s, mobile:%s, body:%s",
		orderId, accountId, name, identity, countryCode, mobile, str)

	if err != nil {
		logs.Error("[CheckAdvanceBlacklist] BlacklistCheck return error, orderId:%d, accountId:%d, name:%s, identity:%s, countryCode:%s, mobile:%s, err:%v",
			orderId, accountId, name, identity, countryCode, mobile, err)
		return "", map[string]bool{}, err
	}

	newM := models.AccountAdvance{}
	newM.Type = 1
	newM.Ctime = tools.GetUnixMillis()
	newM.OrderId = orderId
	newM.AccountId = accountId
	newM.Response = str

	err = newM.Insert()
	if err != nil {
		logs.Error("[CheckAdvanceBlacklist] Insert return error, orderId:%d, accountId:%d, data:%s, err:%v",
			orderId, accountId, str, err)
	}

	return advance.BalcklistPass(&newM)
}

func AdvanceMultiPlatform(orderId, accountId int64, identity string) (ret advance.ResponseData) {
	body, _, err := advance.MultiRecordsCheck(accountId, identity)
	str := string(body)

	logs.Debug("[advanceMultiPlatform] MultiRecordsCheck return orderId:%d, accountId:%d, identity:%s, body:%s",
		orderId, accountId, identity, str)

	if err != nil {
		logs.Error("[advanceMultiPlatform] MultiRecordsCheck return error, orderId:%d, accountId:%d, identity:%s, err:%v",
			orderId, accountId, identity, err)
		return
	}

	err = json.Unmarshal([]byte(str), &ret)
	if err != nil {
		logs.Error("[advanceMultiPlatform] Unmarshal return error, orderId:%d, accountId:%d, identity:%s, err:%v str:%v",
			orderId, accountId, identity, err, str)
	}

	m := models.AccountAdvance{}
	m.Type = 2
	m.Ctime = tools.GetUnixMillis()
	m.OrderId = orderId
	m.AccountId = accountId
	m.Response = str

	err = m.Insert()
	if err != nil {
		logs.Error("[advanceMultiPlatform] Insert return error, orderId:%d, accountId:%d, identity:%s, data:%s, err:%v",
			orderId, accountId, identity, str, err)
		return
	}

	return
}

type PhoneVerifyCallDetails struct {
	Id           int64 `orm:"pk;"`
	OpUid        int64
	PhoneTime    int64
	PhoneConnect int
	Result       int
	Remark       string

	AnswerTimestamp int64
	EndTimestamp    int64
	HangupDirection int
	HangupCause     int
	CallMethod      int
}

func GetPhoneVerifyCallDetailListByOrderIds(orderId int64) (list []PhoneVerifyCallDetails, err error) {

	obj := models.PhoneVerifyCallDetail{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	// 初始化查询条件
	selectSql := fmt.Sprintf(`SELECT op_uid, phone_time, phone_connect, result, remark`)
	where := fmt.Sprintf(`where phone_verify_call_detail.order_id = %v`, orderId)
	sqlList := fmt.Sprintf(`%s FROM %s %s ORDER BY phone_verify_call_detail.id desc`, selectSql, obj.TableName(), where)

	// 查询指定页
	r := o.Raw(sqlList)
	r.QueryRows(&list)

	// 查询'通话记录'
	objSipCallRecord := models.SipCallRecord{}
	o.Using(objSipCallRecord.UsingSlave())

	selectSipCallRecord := `SELECT answer_timestamp, end_timestamp, hangup_direction, hangup_cause, call_method`
	for k, v := range list {
		if v.PhoneTime > 0 {
			var dataSipCallRecord PhoneVerifyCallDetails
			whereSipCallRecord := fmt.Sprintf(`where start_timestamp = %d and call_method = 3`, v.PhoneTime)
			sql := fmt.Sprintf(`%s from %s %s`, selectSipCallRecord, objSipCallRecord.TableName(), whereSipCallRecord)
			r := o.Raw(sql)
			r.QueryRow(&dataSipCallRecord)

			if dataSipCallRecord.CallMethod == voip.VoipCallMethodSipCall {
				list[k].AnswerTimestamp = dataSipCallRecord.AnswerTimestamp
				list[k].EndTimestamp = dataSipCallRecord.EndTimestamp
				list[k].HangupCause = dataSipCallRecord.HangupCause
				list[k].HangupDirection = dataSipCallRecord.HangupDirection
				list[k].CallMethod = dataSipCallRecord.CallMethod
			} else {
				list[k].CallMethod = voip.VoipCallManual
			}
		} else {
			list[k].CallMethod = voip.VoipCallManual
		}
	}

	return
}

type RiskCtlE018Conf struct {
	N1 float64 `json:"N1"`
	N2 int     `json:"N2"`
}

func GetE018Config() (ret RiskCtlE018Conf) {
	riskCtlE018 := config.ValidItemString("risk_ctl_E018")
	err := json.Unmarshal([]byte(riskCtlE018), &ret)
	if err != nil {
		logs.Error("[GetE018Config] Unmarshal riskCtlE018:%s  err:%v", riskCtlE018, err)

		//如果配置错误的话使用缺省值。都通过
		ret.N1 = 1
		ret.N2 = 9999999
	}

	return ret
}

type RiskCtlG003004Config struct {
	Period1To7     int `json:"period_1-7d"`
	Period1To14    int `json:"period_1-14d"`
	Period1To21    int `json:"period_1-21d"`
	Period1To30    int `json:"period_1-30d"`
	Period1To60    int `json:"period_1-60d"`
	Period1To90    int `json:"period_1-90d"`
	Period90ToEver int `json:"period_90+d"`
	Sum            int `json:"sum"`
}

type Info struct {
	FiledName string
	Index     string
}

var RespodColNameMap = map[string]Info{
	"1-7d":  {"Period1To7", "1"},
	"1-14d": {"Period1To14", "2"},
	"1-21d": {"Period1To21", "3"},
	"1-30d": {"Period1To30", "4"},
	"1-60d": {"Period1To60", "5"},
	"1-90d": {"Period1To90", "6"},
	"90+d":  {"Period90ToEver", "7"},
}

func GetRiskCtlG034Config(itemName string) (ret RiskCtlG003004Config) {
	strV := config.ValidItemString(itemName)
	logs.Info("itemName:%s strV:%v", itemName, strV)
	err := json.Unmarshal([]byte(strV), &ret)
	if err != nil {
		logs.Error("[GetRiskCtlG034Config] Unmarshal itemName:%s  err:%v", itemName, err)

		//如果配置错误的话使用缺省值。都通过
		defaultV := 999999
		ret.Period1To7 = defaultV
		ret.Period1To14 = defaultV
		ret.Period1To21 = defaultV
		ret.Period1To30 = defaultV
		ret.Period1To60 = defaultV
		ret.Period1To90 = defaultV
		ret.Period90ToEver = defaultV
		ret.Sum = defaultV
	}
	return
}

func GetConfigValueByColNameV2(config RiskCtlG003004Config, nameCol string) int {
	value := getValueByColNameV2(&config, nameCol).Int()
	return int(value)
}
