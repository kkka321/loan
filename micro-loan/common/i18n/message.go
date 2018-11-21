package i18n

import (
	"github.com/astaxie/beego"

	"micro-loan/common/tools"
)

func GetMessageText(msg int) string {
	region := beego.AppConfig.String("service_region")

	v, ok := regionMessage[region]
	if !ok {
		return ""
	}

	s, ok := v[msg]
	if !ok {
		return ""
	}

	return s
}

const (
	TextRollSuccess = iota

	TextSmsDisburseSuccess
	TextSmsRefundDisburseSuccess

	TextRepayRemind

	TextDefSmsVerify
	TextLoginSmsVerify
	TextLoanSmsVerify

	TextDefVoiceVerify

	TextCollectionRemindDef
	TextCollectionRemindTwo
	TextCollectionRemindFour
	TextCollectionRemindEight

	TextRollApplySuccess

	MsgReviewPassTitle
	MsgReviewPass

	MsgReviewRejectTitle
	MsgReviewInvalidTitle
	MsgReviewReject
	MsgReviewInvalid
	MsgReviewInvalidIdentifyFog
	MsgReviewInvalidIdentifyHoldFog
	MsgReviewInvalidIdentifyHoldNoFace
	MsgReviewInvalidIdentifyHoldNoIdentify

	MsgCreditIncreaseTitle
	MsgCreditIncrease

	MsgLoanSuccessTitle
	MsgLoanSuccess

	MsgLoanFailTitle
	MsgLoanFail

	MsgWaitRepaymentTitle
	MsgWaitRepayment

	MsgRepaymentSuccessTitle
	MsgRepaymentSuccess

	MsgOverdueTitle
	MsgOverdue

	MsgRegisterRemindTitle
	MsgRegisterRemind

	MsgPasswordUnset
	MsgPasswordErr
	MsgUserLocking
	MsgUserLocked

	MsgRollApplySuccessTitle
	MsgRollApplySuccess

	MsgRollSuccessTitle
	MsgRollSuccess

	HomeOrderTagPayedAmount
	HomeOrderTagPenalty
	HomeOrderTagReducedPenalty
	HomeOrderTagMixPayAmount
)

var indonesiaMessage = map[int]string{
	//恭喜您展期成功，应还金额RP %d，%s还款，祝您生活愉快！
	TextRollSuccess: "Pengajuan masa tunda Anda telah diterima. Anda diharapkan untuk melunasi jumlah dana sebanyak Rp%d pada tanggal %s. Terima kasih! ",

	//您好，您在％s上提交的贷款，资金已发送到您的银行帐户，请查看
	TextSmsDisburseSuccess: "Pelanggan Yth,pinjaman diajukan pd %s, dana telah cair k rekening bank Anda, silakan periksa kembali.",

	//资金已退回您的银行帐户，请注意帐户变更。 感谢您的理解和支持。
	TextSmsRefundDisburseSuccess: "Dana telah dikembalikan ke rekening bank Anda, harap perhatikan mutasi rekening. Terima kasih atas pengertian dan dukungan Anda.",

	//您好，您的付款日期是％s，目前的付款金額是Rp％d。 請先安排您的資金。 願你過得愉快
	// TextRepayRemind: "Pelanggan yang terhormat, tanggal pembayaran Anda adalah %s, dan jumlah pembayaran saat ini adalah Rp%d. Silakan mengatur dana Anda terlebih dahulu. Semoga Anda hidup bahagia!",
	//`尊敬的客户，您的付款日期是{3.24}，sbsr Rp {20000}. 将您的资金设置为dhl. Smg hdp bhgia！ VA:{BNI 1234567891234567}`
	TextRepayRemind: `Pelanggan Yth, tgl pembayaran Anda adlh %s, sbsr Rp%d. Slkn atur dana Anda trlebih dhl. Smg hdp bhgia! VA:%s`,

	//您的验证码％s，有效期为10分钟，请尽快核实
	TextDefSmsVerify: "Kode verifikasi Anda %s, berlaku selama 10 menit, mohon verifikasi sesegera mungkin.",
	//您的验证码％s，有效期为10分钟，请尽快核实
	TextLoginSmsVerify: "Kode verifikasi Anda %s, berlaku selama 10 menit, mohon verifikasi sesegera mungkin.",
	//您的验证码％s，有效期为10分钟，请尽快核实
	TextLoanSmsVerify: "Kode verifikasi Anda %s, berlaku selama 10 menit, mohon verifikasi sesegera mungkin.",

	//您的验证码是:-%s
	TextDefVoiceVerify: "Kode verifikasi adalah:-%s",

	//您的贷款过期％dhr，ttl返回％d，tlng skrng返回，以免object.VA％s
	TextCollectionRemindDef:   "RupiahCepat:Pinjaman anda sdh jatuh tempo %dhr,ttl pengembalian %d,tlng skrng kembalianny agar tdk keberatan,%s VA%s",
	TextCollectionRemindTwo:   "RupiahCepat:Pinjaman anda sdh jatuh tempo %dhr,ttl pengembalian %d,tlng skrng kembalianny agar tdk keberatan,%s VA%s",
	TextCollectionRemindFour:  "RupiahCepat:Pinjaman anda sdh jatuh tempo %dhr,ttl pengembalian %d,tlng skrng kembalianny agar tdk keberatan,%s VA%s",
	TextCollectionRemindEight: "RupiahCepat:Anda sdh keterlambatan %dhr,ttl tagihan jd %d,tlng kembali segera agr tdk ganggu kredit selnjtny,%s VA%s",

	//您已申请展期成功，最低还款额xxxx，请于今天还款，过期失效，%银行 VA %账户，谢谢!
	TextRollApplySuccess: `Pengajuan masa tunda telah disetujui,mohon bayar minimal Rp%d,pada hari ini,lewat waktu akan kedaluwarsa,%s VA%s,Trmksh`,

	//`订单审核通过`
	MsgReviewPassTitle: `Pengajuan Pinjaman telah disetujui`,
	//`订单审核通过，请进入“我的账户”页面查看详情`
	MsgReviewPass: `Pinjaman sudah disetujui, kami akan segera mengatur pengiriman dana, Anda bisa melihat detailnya di halaman "Akun Saya".`,

	//`订单审核拒绝`
	MsgReviewRejectTitle: `Pengajuan Pinjaman Ditolak`,
	//`抱歉，您的订单审核未通过，请进入“我的账户”页面查看详情`
	MsgReviewReject: `Maaf, pengkajian pengajuan pinjaman gagal, Anda bisa melihat detailnya di halaman "Akun Saya".`,

	//`订单置为无效`
	MsgReviewInvalidTitle: `Pemberitahuan Pengajuan Pinjaman Gagal`,

	//` 经检查，您上传的KTP/手持KTP 不符合要求，请重新上传！`
	MsgReviewInvalid: `Setelah diperiksa, KTP / KTP genggam yang Anda unggah tidak memenuhi persyaratan, silakan unggah lagi!".`,

	//  身份证照片模糊，身份证信息无法识
	MsgReviewInvalidIdentifyFog: `Pengajuan Anda gagal karena foto KTP Anda kabur atau tidak jelas sehingga tidak dapat memverifikasi informasi KTP Anda. Silakan ajukan sekali lagi dan unggah foto KTP sesuai petunjuk.`,

	// 手持证件照模糊
	MsgReviewInvalidIdentifyHoldFog: `Pengajuan Anda gagal karena foto KTP yang Anda pegang kabur atau tidak jelas sehingga tidak dapat memverifikasi informasi KTP Anda. Silakan ajukan sekali lagi dan mengambil foto Anda beserta KTP Anda sesuai dengan petunjuk.`,

	//  手持证件照中缺少人脸
	MsgReviewInvalidIdentifyHoldNoFace: `Pengajuan Anda gagal karena wajah Anda tidak terdeteksi atau tidak terlihat jelas saat memegang KTP sehingga tidak dapat memverifikasi informasi KTP Anda. Silakan ajukan sekali lagi dan ikuti petunjuk untuk mengambil foto Anda beserta KTP Anda.`,

	//  手持证件照中缺少身份证
	MsgReviewInvalidIdentifyHoldNoIdentify: `Pengajuan Anda gagal karena kartu identitas yang Anda pegang bukan KTP sehingga tidak dapat memverifikasi informasi KTP Anda. Silakan ajukan sekali lagi dan ikuti petunjuk dalam pengambilan foto Anda beserta KTP Anda.`,

	//恭喜您增加额度
	MsgCreditIncreaseTitle: `Selamat, limit Anda telah meningkat!`,

	//您的授已成功，可在首页侧边栏授信信息查看
	MsgCreditIncrease: `Akun Anda berhasil diverifikasi, update app ke versi terbaru untuk melihat detailnya di menu Menaikkan Skor Kredit.`,

	//`放款成功提醒`
	MsgLoanSuccessTitle: `Peringatan Pengiriman Dana Berhasil`,
	//`借款资金已汇入您的银行账户，请核实。`
	MsgLoanSuccess: `Dana yang dipinjam telah ditransfer ke rekening bank Anda. Harap dikonfirmasi.`,

	//`放款失败提醒`
	MsgLoanFailTitle: `Peringatan Pengiriman Dana Gagal`,
	//`抱歉，资金汇入您的银行账户失败，请联系客服人员进行处理。`
	MsgLoanFail: `Maaf, dana gagal dikirim ke akun bank Anda. Harap menghubungi petugas layanan pelanggan untuk memproses. Saluran siaga pelanggan: 085825100001`,

	//`还款提醒`
	MsgWaitRepaymentTitle: `Peringatan Pengembalian`,
	//`您的借款即将到期，请您及时还款以免影响您的信用。如您对如何还款有疑问，请查看“还款指引”页面。`
	MsgWaitRepayment: ` Pinjaman Anda hampir kedaluwarsa. Harap melunasi tagihan Anda tepat waktu untuk menjaga kredit Anda. Jika ada pertanyaan, harap periksa halaman "CARA PEMBAYARAN".`,

	//`还款成功提醒`
	MsgRepaymentSuccessTitle: `Peringatan Pelunasan Berhasil`,
	//`您的借款已还清，感谢您对Rupiah Cepat的支持 。我们任何的不足之处，您可进入“反馈”页面帮助我们改善。`
	MsgRepaymentSuccess: `Tagihan Anda telah dilunasi. Terima kasih atas dukungan Anda. Selamat datang di halaman "Umpan Balik" untuk membantu kami meningkatkan.`,

	//`逾期提醒`
	MsgOverdueTitle: `Peringatan Keterlambatan Pengembalian`,
	//`您的借款已逾期，请您尽快还款以免影响您的信用。如您对如何还款有疑问，请查看“还款指引”页面。`
	MsgOverdue: `Pinjaman Anda sudah kedaluwarsa. Harap membayar tagihan sesegera mungkin supaya tidak mempengaruhi kredit Anda. Jika ada pertanyaan, harap periksa halaman "CARA PEMBAYARAN".`,

	//`【Rupiah Cepat】信用当钱花`
	MsgRegisterRemindTitle: `【Rupiah Cepat】Kredit yang cepat`,
	//`还在为借不到钱烦恼吗？Rupiah Cepat来帮你，信用当钱花！戳我注册！`
	MsgRegisterRemind: `Masih khawatir tak bisa pinjam uang? Rupiah Cepat hadir di sini untuk membantu masalah keuangan anda, kredit yang cepat! Klik di sini untuk mendaftar!`,

	//`密码未设置`
	MsgPasswordUnset: `Anda belum mempunyai kata sandi, silahkan masuk dengan SMS dan mengatur kata sandi Anda.`,
	//`密码错误`
	MsgPasswordErr: `Kata sandi salah`,
	//`密码再错6-n次账户将被锁定，账号锁定24小时再可再次登录`
	MsgUserLocking: "Anda dapat masuk setelah 24 jam kalau password salah %d kali lagi.",
	//`密码输错达6次，账号锁定24小时再可再次登录`
	MsgUserLocked: "Password telah salah sampai %d kali, Anda dapat coba ulang setelah 24 jam.",

	MsgRollApplySuccessTitle: `Pemberitahuan keberhasilan masa tunda`, //展期申请成功通知
	//您已申请展期成功，最低还款额xxxx，请于%最晚还款时间%前还款，过期失效，%银行 VA %账户，谢谢！
	MsgRollApplySuccess: `Pengajuan masa tunda Anda telah disetujui, mohon kembalikan dana minimal Rp%d sebelum %s, kalau lewat waktu akan kedaluwarsa, %s VA%s,terima kasih!`,

	MsgRollSuccessTitle: `Pemberitahuan keberhasilan masa tunda`, //展期成功通知
	//您已展期成功，展期后还款金额为%剩余未还总额，还款日期xxxx，展期期间不再计罚息，请注意还款，%银行 VA %账户,祝您生活愉快!
	MsgRollSuccess: `Masa tunda Anda sedang berjalan, mohon kembalikan jumlah dana Rp%d sebelum %s, Anda tidak terkena bunga denda selama masa tunda, mohon bayar tepat waktu, %s VA%s,terima kasih!`,

	// 首页显示在贷订单的信息
	HomeOrderTagPayedAmount:    "Dana telah dikembalikan", // 已还金额
	HomeOrderTagPenalty:        "Bunga denda",             // 罚息
	HomeOrderTagReducedPenalty: "Bunga denda dikurangi",   // 已减免罚息
	HomeOrderTagMixPayAmount:   "Pengembalian minimal",    // 最低还款金额
}

var indiaMessage = map[int]string{
	TextRollSuccess:    "恭喜您展期成功，应还金额RP %d，%s还款，祝您生活愉快！",
	TextLoginSmsVerify: "Your login verification code is %s, valid for 10 minutes, please do not reveal it to anyone.",
}

var regionMessage = map[string](map[int]string){
	tools.ServiceRegionIndia:     indiaMessage,
	tools.ServiceRegionIndonesia: indonesiaMessage,
}
