package i18n

// 可以灵活配置,实现多语言包功能
var langMap = map[string]map[string]string{
	"新增": LangItem{
		LangEnUS: "add",
		LangIdID: "tambah",
	},
	"角色名": LangItem{
		LangEnUS: "role name",
		LangIdID: "nama akun",
	},
	"角色管理": LangItem{
		LangEnUS: "role management",
		LangIdID: "pengelola akun",
	},
	"权限管理": LangItem{
		LangEnUS: "authority management",
		LangIdID: "pengelola hak",
	},
	// 分页相关 {{{
	"首页": LangItem{
		LangEnUS: "First",
		LangIdID: "halaman pertama",
	},

	"后台系统": LangItem{
		LangEnUS: "backend system",
		LangIdID: "sistem admin",
	},

	"欢迎进入后台系统": LangItem{
		LangEnUS: "Welcome to backend",
		LangIdID: "selamat datang di sistem administrasi Rupiah Cepat",
	},

	"尾页": LangItem{
		LangEnUS: "Tail",
		LangIdID: "halaman terakhir",
	},
	// }}}
	"RiskCtl Manage": LangItem{
		LangZhCN: "风险控制管理",
		LangIdID: "pengelola pengawasan resiko",
	},
	"Risk List": LangItem{
		LangZhCN: "风控列表",
		LangIdID: "daftar pengawasan resiko",
	},
	"产品管理": LangItem{
		LangEnUS: "Product Management",
		LangIdID: "pengelolaan produk",
	},
	"客户风险管理": LangItem{
		LangEnUS: "Customer risk management",
		LangIdID: "pengawasan resiko nasabah",
	},
	"产品列表": LangItem{
		LangEnUS: "Product List",
		LangIdID: "daftar produk",
	},
	"添加产品": LangItem{
		LangEnUS: "Add Product",
		LangIdID: "menambah produk",
	},
	"编辑产品": LangItem{
		LangEnUS: "Edit Product",
		LangIdID: "edit produk",
	},
	"金融产品": LangItem{
		LangEnUS: "Financial product",
		LangIdID: "produk finansial",
	},
	"产品名称": LangItem{
		LangEnUS: "Product Name",
		LangIdID: "nama produk",
	},
	"版本": LangItem{
		LangEnUS: "version",
		LangIdID: "versi",
	},
	"状态": LangItem{
		LangEnUS: "Status",
		LangIdID: "Status",
	},
	"日利率": LangItem{
		LangEnUS: "Daily rate",
		LangIdID: "bunga harian",
	},
	"利息类型": LangItem{
		LangEnUS: "Interest type",
		LangIdID: "tipe bunga",
	},
	"日费率": LangItem{
		LangEnUS: "Daily fee rate",
		LangIdID: "bunga dan admin harian",
	},
	"费用类型": LangItem{
		LangEnUS: "fee type",
		LangIdID: "tipe biaya",
	},
	"宽限期利率": LangItem{
		LangEnUS: "grace period interest rate",
		LangIdID: "Bunga masa tenggang",
	},
	"罚息利率": LangItem{
		LangEnUS: "penalize interest rate",
		LangIdID: "Bunga denda",
	},
	"最小金额": LangItem{
		LangEnUS: "Minimum amount",
		LangIdID: "nominal minimal",
	},
	"最大金额": LangItem{
		LangEnUS: "Maximum amount",
		LangIdID: "nominal maksimal",
	},
	"最短期限": LangItem{
		LangEnUS: "Minimum period",
		LangIdID: "toner minimal",
	},
	"最长期限": LangItem{
		LangEnUS: "Maximum period",
		LangIdID: "toner maksimal",
	},
	"期限单位": LangItem{
		LangEnUS: "term unit",
		LangIdID: "jangka waktu",
	},
	"日": LangItem{
		LangEnUS: "day",
		LangIdID: "hari",
	},

	"取整方式": LangItem{
		LangEnUS: "rounding method",
		LangIdID: "Cara pembulatan",
	},
	"取整单位": LangItem{
		LangEnUS: "rounding units",
		LangIdID: "Unit pembulatan",
	},
	"宽限期": LangItem{
		LangEnUS: "grace period",
		LangIdID: "Masa tenggang",
	},
	"还款顺序": LangItem{
		LangEnUS: "repayment order",
		LangIdID: "Urutan pengembalian",
	},
	"未生效": LangItem{
		LangEnUS: "not active",
		LangIdID: "tidak efektif",
	},
	"放款时扣取": LangItem{
		LangEnUS: "deduction when lending money",
		LangIdID: "Dipotong saat cair dana",
	},
	"分期还": LangItem{
		LangEnUS: "instalments",
		LangIdID: "cicil",
	},
	"向上取整": LangItem{
		LangEnUS: "rounded up",
		LangIdID: "pembulatan",
	},
	"不取整": LangItem{
		LangEnUS: "no rounding",
		LangIdID: "Tidak pembulatan",
	},
	"试算": LangItem{
		LangEnUS: "trial",
		LangIdID: "Percobaan",
	},
	"克隆": LangItem{
		LangEnUS: "clone",
		LangIdID: "Kloning",
	},

	"宽限期日费率": LangItem{
		LangEnUS: "Daily fee rate of grace period",
		LangIdID: "Biaya dan Bunga masa tenggang",
	},
	"罚息日费率": LangItem{
		LangEnUS: "Daily fee rate of penalty",
		LangIdID: "Biaya dan Bunga denda",
	},
	"利息收取类型": LangItem{
		LangEnUS: "interest type",
		LangIdID: "Jenis Bungan",
	},
	"费用收取类型": LangItem{
		LangEnUS: "charge type",
		LangIdID: "Jenis biaya",
	},
	"最小借款金额": LangItem{
		LangEnUS: "min loan amount",
		LangIdID: "Pinjaman minimal",
	},
	"最大借款金额": LangItem{
		LangEnUS: "max loan amount",
		LangIdID: "Pinjaman maksimal",
	},
	"最短借款期限": LangItem{
		LangEnUS: "shortest loan period",
		LangIdID: "Tenor minimal",
	},
	"最长借款期限": LangItem{
		LangEnUS: "longest loan period",
		LangIdID: "Tenor maksimal",
	},

	"还款方式": LangItem{
		LangEnUS: "pay method",
		LangIdID: "Cara pengembalian",
	},
	"暂时只实现了一次性还本付息": LangItem{
		LangEnUS: "Only support principal and interest paid off once",
		LangIdID: "Sementara hanya mendukung sekalian kembali semuanya",
	},
	"一次性还本付息": LangItem{
		LangEnUS: "Principal and interest paid off once",
		LangIdID: "Sekalian kembali tunggakan pokok, bunga dan biaya administrasi",
	},
	"部分还款顺序": LangItem{
		LangEnUS: "Partial repayment order",
		LangIdID: "Urutan pengambalian sebagian",
	},
	"滞纳金": LangItem{
		LangEnUS: "Late payment",
		LangIdID: "Forfeit",
	},
	"使用": LangItem{
		LangEnUS: "use",
		LangIdID: "guna",
	},
	"分割": LangItem{
		LangEnUS: "split",
		LangIdID: "pisah",
	},
	"如": LangItem{
		LangEnUS: "example",
		LangIdID: "contoh",
	},

	"产品试算": LangItem{
		LangEnUS: "Product trial",
		LangIdID: "Kalkulator",
	},
	"试算字段": LangItem{
		LangEnUS: "Trial field",
		LangIdID: "Bidang percobaan",
	},
	"试算结果": LangItem{
		LangEnUS: "Trial result",
		LangIdID: "Hasil uji coba",
	},
	"期数": LangItem{
		LangEnUS: "period",
		LangIdID: "periode",
	},
	"产品": LangItem{
		LangEnUS: "Product",
		LangIdID: "produk",
	},
	"放款日期": LangItem{
		LangEnUS: "loan date",
		LangIdID: "tgl cair dana",
	},
	"当前日期": LangItem{
		LangEnUS: "current date",
		LangIdID: "tgl skrng",
	},
	"已还总额": LangItem{
		LangEnUS: "total paid",
		LangIdID: "total sudah dikembali",
	},
	"用户标识": LangItem{
		LangEnUS: "user identification",
		LangIdID: "status nasabah",
	},
	"应还滞纳金": LangItem{
		LangEnUS: "late fees unpaid",
		LangIdID: "hangus yg harus dibalikan",
	},
	"已还滞纳金": LangItem{
		LangEnUS: "late fees paid",
		LangIdID: "hangus yg udah dibalikan",
	},

	"还款提醒": LangItem{
		LangEnUS: "Repayment reminder",
		LangIdID: "ingatkan kembalian",
	},
	"逾期提醒": LangItem{
		LangEnUS: "Overdue reminder",
		LangIdID: "ingatkan keterlambatan",
	},
	"罚息公式": LangItem{
		LangEnUS: "Penalty formula",
		LangIdID: "remus denda",
	},
	"金额": LangItem{
		LangEnUS: "Amount",
		LangIdID: "nomimal",
	},
	"时间": LangItem{
		LangEnUS: "time",
		LangIdID: "waktu",
	},
	"操作": LangItem{
		LangEnUS: "Operations",
		LangIdID: "Operasi",
	},
	"操作员": LangItem{
		LangEnUS: "Operator",
		LangIdID: "operator",
	},
	"操作时间": LangItem{
		LangEnUS: "Operate time",
		LangIdID: "waktu operasi",
	},
	"客户": LangItem{
		LangEnUS: "Customer",
		LangIdID: "nasabah",
	},
	"系统": LangItem{
		LangEnUS: "System",
		LangIdID: "sistem",
	},
	"实还日期": LangItem{
		LangEnUS: "Actual date",
		LangIdID: "tanggal kembalian asli",
	},
	"本金": LangItem{
		LangEnUS: "Principal",
		LangIdID: "tunggakan",
	},
	"砍头息": LangItem{
		LangEnUS: "Prefee",
		LangIdID: "tunggakan",
	},
	"宽限期利息": LangItem{
		LangEnUS: "Grace period interest",
		LangIdID: "bunga period grace",
	},
	"服务费": LangItem{
		LangEnUS: "Service fee",
		LangIdID: "biaya admin",
	},
	"应还本金": LangItem{
		LangEnUS: "Should pay principal",
		LangIdID: "tunggakan wajib balik",
	},
	"已还本金": LangItem{
		LangEnUS: "Repayment principal",
		LangIdID: "tunggakan sudah dibalik",
	},
	"应还利息": LangItem{
		LangEnUS: "Should pay interest",
		LangIdID: "bunga wajib kembalian",
	},
	"已还利息": LangItem{
		LangEnUS: "Repayment Interest",
		LangIdID: "bunga yang sudah dibalik",
	},
	"应还服务费": LangItem{
		LangEnUS: "Should service fee",
		LangIdID: "biaya admin wajib dibalik",
	},
	"已还服务费": LangItem{
		LangEnUS: "Repayment service fee",
		LangIdID: "biaya admin sudah dikembalian",
	},
	"罚息": LangItem{
		LangEnUS: "Penalty interest",
		LangIdID: "bunga denda",
	},
	"应还罚息": LangItem{
		LangEnUS: "Should penalize interest",
		LangIdID: "bunga denda wajib dibalik",
	},
	"已还罚息": LangItem{
		LangEnUS: "Repayment penalize interest",
		LangIdID: "bunga denda sudah dibalik",
	},
	"提交借款申请": LangItem{
		LangEnUS: "Submit loan application",
		LangIdID: "ajukan permohonan pinjaman",
	},
	"提交借款审核": LangItem{
		LangEnUS: "Submit loan review",
		LangIdID: "ajukan pemeriksa pinjaman",
	},
	"反欺诈规则": LangItem{
		LangEnUS: "Risk control regular",
		LangIdID: "regulasi anti-penipuan",
	},
	"电核完成": LangItem{
		LangEnUS: "Phone verify completion",
		LangIdID: "verifikasi berhasil",
	},
	"人工审核通过": LangItem{
		LangEnUS: "Manual review passed",
		LangIdID: "lulus varifikasi manual",
	},
	"评分": LangItem{
		LangEnUS: "Score",
		LangIdID: "Score",
	},
	"评分模型": LangItem{
		LangEnUS: "Scoring model",
		LangIdID: "Scoring model",
	},
	"编辑": LangItem{
		LangEnUS: "Edit",
		LangIdID: "Edit",
	},
	"上架": LangItem{
		LangEnUS: "Enable",
		LangIdID: "Enable",
	},
	"下架": LangItem{
		LangEnUS: "Disable",
		LangIdID: "Disable",
	},
	"客户管理": LangItem{
		LangEnUS: "Customer Management",
		LangIdID: "pengelola nasabah",
	},
	"客户列表": LangItem{
		LangEnUS: "Customer List",
		LangIdID: "daftar nasabah",
	},
	"风险管理": LangItem{
		LangEnUS: "Risk Management",
		LangIdID: "pengawasan resiko",
	},
	"账号": LangItem{
		LangEnUS: "AccountID",
		LangIdID: "nasabah",
	},
	"客户姓名": LangItem{
		LangEnUS: "Name",
		LangIdID: "nama",
	},
	"手机号": LangItem{
		LangEnUS: "Phone No",
		LangIdID: "nomor hp",
	},
	"身份认证状态": LangItem{
		LangEnUS: "KTP Certification",
		LangIdID: "status varifikasi KTP",
	},
	"客户分类": LangItem{
		LangEnUS: "Customer",
		LangIdID: "katerogi nasabah",
	},
	"当前额度": LangItem{
		LangEnUS: "Current Credit",
		LangIdID: "kredit skrng",
	},
	"放款金额": LangItem{
		LangEnUS: "Loan Amount",
		LangIdID: "jumlah disetuju",
	},
	"改订单状态": LangItem{
		LangEnUS: "Change Status",
		LangIdID: "ubah status",
	},
	"来源渠道": LangItem{
		LangEnUS: "Source",
		LangIdID: "sumber data",
	},
	"尚未识别": LangItem{
		LangEnUS: "Not Identified",
		LangIdID: "belum ketahuan",
	},

	"暂未实现": LangItem{
		LangEnUS: "Not Implement",
		LangIdID: "belum ketahuan",
	},

	"注册时间": LangItem{
		LangEnUS: "Registration Time",
		LangIdID: "waktu registrasi",
	},
	"沟通": LangItem{
		LangEnUS: "communication",
		LangIdID: "komunikasi",
	},
	"上传时间": LangItem{
		LangEnUS: "Upload Time",
		LangIdID: "Waktu pengunggaha",
	},
	"图片用途": LangItem{
		LangEnUS: "Picture Use",
		LangIdID: "Penggunaan gambar",
	},
	"图片": LangItem{
		LangEnUS: "Picture",
		LangIdID: "Gambar",
	},
	"图片历史": LangItem{
		LangEnUS: "Picture History",
		LangIdID: "Sejarah gambar",
	},
	"上报风险": LangItem{
		LangEnUS: "Report Risk",
		LangIdID: "lapor resiko",
	},
	"客户沟通": LangItem{
		LangEnUS: "Customer communication",
		LangIdID: "komunikasi dengan nasabah",
	},
	"沟通时间": LangItem{
		LangEnUS: "Communication time",
		LangIdID: "waktu komunikasi",
	},
	"沟通人": LangItem{
		LangEnUS: "Communicator",
		LangIdID: "orang komunikasi",
	},
	"沟通内容": LangItem{
		LangEnUS: "Contents",
		LangIdID: "isi",
	},
	// 客户标签 {{{
	"潜在客户": LangItem{
		LangEnUS: "Potential customers",
		LangIdID: "nasabah potensi",
	},
	"目标客户": LangItem{
		LangEnUS: "Target customers",
		LangIdID: "nasabah target",
	},
	"准客户": LangItem{
		LangEnUS: "Prospective customers",
		LangIdID: "calon nasabah",
	},
	"成交客户": LangItem{
		LangEnUS: "Deal customers",
		LangIdID: "nasabah terjadi",
	},
	"忠实客户": LangItem{
		LangEnUS: "Loyal customers",
		LangIdID: "nasabah setia",
	},
	// }}}
	"上报": LangItem{
		LangEnUS: "Report",
		LangIdID: "lapor",
	},
	"解除": LangItem{
		LangEnUS: "Release",
		LangIdID: "lepas",
	},
	"解除原因": LangItem{
		LangEnUS: "Release Reason",
		LangIdID: "alasan pelepasan",
	},
	"风险": LangItem{
		LangEnUS: "Risk",
		LangIdID: "resiko",
	},
	"原因": LangItem{
		LangEnUS: "Reason",
		LangIdID: "alasan",
	},
	"说明": LangItem{
		LangEnUS: "Note",
		LangIdID: "catatan",
	},
	"请填写": LangItem{
		LangEnUS: "Please fill out",
		LangIdID: "silahkan mengisi",
	},
	// 上报原因相关 {{{
	"中介代办": LangItem{
		LangEnUS: "Intermediary agency",
		LangIdID: "diwakili agent",
	},
	"欺诈": LangItem{
		LangEnUS: "Fraud",
		LangIdID: "pengipu",
	},
	"负面信息": LangItem{
		LangEnUS: "Negative information",
		LangIdID: "informasi negatif",
	},
	"同业参与": LangItem{
		LangEnUS: "Industry participation",
		LangIdID: "Partisipasi industri",
	},
	"手机号码": LangItem{
		LangEnUS: "Mobile number",
		LangIdID: "nomor hp",
	},
	"身份证号码": LangItem{
		LangEnUS: "Identification number",
		LangIdID: "NIK",
	},
	"居住地址": LangItem{
		LangEnUS: "Residence address",
		LangIdID: "domisili",
	},
	"单位名称": LangItem{
		LangEnUS: "Company name",
		LangIdID: "nama PT.",
	},
	"单位地址": LangItem{
		LangEnUS: "Company address",
		LangIdID: "alamt PT.",
	},
	"设备号": LangItem{
		LangEnUS: "Device IMEI",
		LangIdID: "IMEI Perangkat",
	},
	"IP地址": LangItem{
		LangEnUS: "IP address",
		LangIdID: "alamt IP",
	},
	"伪冒申请": LangItem{
		LangEnUS: "Fake others application",
		LangIdID: "apply anti-palsu",
	},
	"组团骗贷": LangItem{
		LangEnUS: "Group fraud loans",
		LangIdID: "Group fraud loans",
	},
	"贷后高风险": LangItem{
		LangEnUS: "High risk after loan",
		LangIdID: "resiko tinggal setelah pinjaman",
	},
	"黑名单管理": LangItem{
		LangEnUS: "Black List Management",
		LangIdID: "Group fraud loans",
	},
	"等待审核": LangItem{
		LangEnUS: "Wait Review",
		LangIdID: "tunggu varifikasi",
	},
	"审核通过": LangItem{
		LangEnUS: "Review Passed",
		LangIdID: "varifikasi berhasil",
	},
	"上报时间": LangItem{
		LangEnUS: "Report Time",
		LangIdID: "waktu laporan",
	},
	"上报说明": LangItem{
		LangEnUS: "Report Note",
		LangIdID: "catatan laporan",
	},
	"审核说明": LangItem{
		LangEnUS: "Review Note",
		LangIdID: "penjelasan periksa",
	},
	"解除说明": LangItem{
		LangEnUS: "Review Note",
		LangIdID: "penjelasan pelepas",
	},
	"审核时间": LangItem{
		LangEnUS: "Review Time",
		LangIdID: "waktu varifikasi",
	},
	"审核状态": LangItem{
		LangEnUS: "Review Status",
		LangIdID: "status varifikasi",
	},
	"内部提报": LangItem{
		LangEnUS: "Internal reporting",
	},
	"系统识别": LangItem{
		LangEnUS: "System",
		LangIdID: "Identifikasi sistem",
	},
	"失联客户": LangItem{
		LangIdID: "kehilangan jaringan user",
	},
	"Akulaku黑名单": LangItem{
		LangIdID: "blacklist Akulaku",
	},
	"选择日期范围": LangItem{
		LangEnUS: "Select date range",
		LangIdID: "pilih tanggal",
	},
	"所有": LangItem{
		LangEnUS: "All",
		LangIdID: "semua",
	},
	"未解除": LangItem{
		LangEnUS: "Not Relieve",
		LangIdID: "belum terlepas",
	},
	"已解除": LangItem{
		LangEnUS: "Already Relieved",
		LangIdID: "sudah terlepas",
	},
	"是否解除": LangItem{
		LangEnUS: "Relieved status",
		LangIdID: "status pelepasan",
	},
	"修正值": LangItem{
		LangEnUS: "Fixed Value",
		LangIdID: "Fixed Value",
	},
	"起始值": LangItem{
		LangEnUS: "Start Value",
		LangIdID: "value pemula",
	},
	"终止值": LangItem{
		LangEnUS: "End Value",
		LangIdID: "value berhenti",
	},
	"反欺诈完成时间": LangItem{
		LangEnUS: "Anti-fraud Finish Time",
		LangIdID: "waktu selesai anti-fraud",
	},
	"电核完成时间": LangItem{
		LangEnUS: "Phone Verify Finish Time",
		LangIdID: "waktu selesai Verifikasi telepon",
	},
	"放款总金额": LangItem{
		LangEnUS: "Total Payment",
		LangIdID: "Total dana cairan",
	},
	"放款成功总金额": LangItem{
		LangEnUS: "Total Payment Success",
		LangIdID: "dana cair berhasil",
	},
	"应还款总金额": LangItem{
		LangEnUS: "Total return",
		LangIdID: "jumlah kembalian",
	},
	"实际还款总金额": LangItem{
		LangEnUS: "Total Amount",
		LangIdID: "nominal kembalian asli",
	},
	"减免总金额": LangItem{
		LangEnUS: "Total Reduction",
		LangIdID: "jumlah dikurangi",
	},
	"实际还款时间": LangItem{
		LangEnUS: "Actual repay time",
		LangIdID: "waktu asli pengembalikan",
	},
	"拒贷码": LangItem{
		LangEnUS: "Repellent Code",
		LangIdID: "code tolak",
	},
	"客户帐号": LangItem{
		LangEnUS: "Customer Account",
		LangIdID: "rekning nasabah",
	},
	"审核": LangItem{
		LangEnUS: "Review",
		LangIdID: "periksa",
	},
	"身份证照片": LangItem{
		LangEnUS: "Identification photo",
		LangIdID: "periksa",
	},
	"活体识别": LangItem{
		LangEnUS: "Live recognition",
		LangIdID: "Live recognition",
	},
	"联系人1手机号": LangItem{
		LangEnUS: "Contact 1 phone number",
		LangIdID: "no hp PIC1",
	},
	"联系人2手机号": LangItem{
		LangEnUS: "Contact 2 phone number",
		LangIdID: "no hp PIC2",
	},
	"风险识别错误": LangItem{
		LangEnUS: "Risk discern error",
		LangIdID: "kesalahan bedakan resiko",
	},
	"负面信息消除": LangItem{
		LangEnUS: "Negative information elimination",
		LangIdID: "hapus informasi negatif",
	},
	// }}}
	"风险类型": LangItem{
		LangEnUS: "Risk Type",
		LangIdID: "tipe resiko",
	},
	"风险项": LangItem{
		LangEnUS: "Risk Item",
		LangIdID: "item resiko",
	},
	"此栏目为必填项": LangItem{
		LangEnUS: "The field is required",
		LangIdID: "wajib isi",
	},
	"风险值": LangItem{
		LangEnUS: "Risk Value",
		LangIdID: "value resiko",
	},
	"加入时间": LangItem{
		LangEnUS: "Join Time",
		LangIdID: "waktu gabung",
	},
	"黑名单": LangItem{
		LangEnUS: "Blacklist",
		LangIdID: "Blacklist",
	},
	"灰名单": LangItem{
		LangEnUS: "Grey list",
		LangIdID: "Grey list",
	},
	// 风险状态 {{{
	"风险状态": LangItem{
		LangEnUS: "Risk Status",
		LangIdID: "status resiko",
	},
	"反欺诈处理中": LangItem{
		LangEnUS: "Anti-fraud processing",
		LangIdID: "sedang proses Anti-fraud",
	},
	"反欺诈拒绝": LangItem{
		LangEnUS: "Anti-fraud reject",
		LangIdID: "ditolak Anti-fraud",
	},
	"反欺诈直批": LangItem{
		LangEnUS: "Anti-fraud direct approval",
		LangIdID: "persetujuan Anti-fraud direct approval",
	},
	"电核": LangItem{
		LangEnUS: "Phone Verify",
		LangIdID: "verifikasi",
	},
	"等待电核": LangItem{
		LangEnUS: "Wait Phone Verify",
		LangIdID: "menunggu varifikasi via telepon",
	},
	"电核处理中": LangItem{
		LangEnUS: "Phone verify processing",
		LangIdID: "proses varifikasi via telepon",
	},
	"电核通过": LangItem{
		LangEnUS: "Phone Verify Pass",
		LangIdID: "lulus varifikasi via telepon",
	},
	"电核拒绝": LangItem{
		LangEnUS: "Phone Verify Reject",
		LangIdID: "ditolak varifikasi via telepon",
	},
	"电核结论": LangItem{
		LangEnUS: "Result",
		LangIdID: "kesimpulan",
	},
	"查看电核结论": LangItem{
		LangEnUS: "View Verify Result",
		LangIdID: "kesimpulan",
	},
	// }}}
	"借款订单编号": LangItem{
		LangEnUS: "Loan ID",
		LangIdID: "ID pinjaman",
	},
	"借款金额": LangItem{
		LangEnUS: "Loan amount",
		LangIdID: "jumlah pinjaman",
	},
	"放贷金额": LangItem{
		LangEnUS: "Payment amount",
		LangIdID: "jumlah dana cair",
	},
	"借款期限": LangItem{
		LangEnUS: "Loan terms",
		LangIdID: "toner pinjaman",
	},
	"风控状态": LangItem{
		LangEnUS: "Risk Status",
		LangIdID: "status pengawas resiko",
	},
	"风控管理": LangItem{
		LangEnUS: "risk control management",
		LangIdID: "resiko manajer",
	},
	"申请时间": LangItem{
		LangEnUS: "Apply Time",
		LangIdID: "waktu ajukan",
	},
	"审批时间": LangItem{
		LangEnUS: "Processing time",
		LangIdID: "waktu proses",
	},
	"电核人员": LangItem{
		LangEnUS: "Auditors",
		LangIdID: "pengawas",
	},
	"暂无": LangItem{
		LangEnUS: "none",
		LangIdID: "belum ada",
	},
	"借款": LangItem{
		LangEnUS: "Loan",
		LangIdID: "pinjaman",
	},
	"借款管理": LangItem{
		LangEnUS: "Loan Management",
		LangIdID: "pengelola pinjaman",
	},
	"放款": LangItem{
		LangEnUS: "payment",
		LangIdID: "dana cair",
	},
	"放款管理": LangItem{
		LangEnUS: "Payment Management",
		LangIdID: "pengelola dana cair",
	},
	"放款状态": LangItem{
		LangEnUS: "Payment Status",
		LangIdID: "status dana cair",
	},
	"修改客户基本信息": LangItem{
		LangIdID: "Ubah data basic nasabah",
	},
	"银行名称": LangItem{
		LangEnUS: "Bank Name",
		LangIdID: "nama bank",
	},
	"银行卡号": LangItem{
		LangEnUS: "Bank No",
		LangIdID: "nomor rekning",
	},
	"备注": LangItem{
		LangEnUS: "Note ",
		LangIdID: "catatan",
	},
	"放款订单生成时间": LangItem{
		LangEnUS: "Create Time",
		LangIdID: "waktu orderan dana cair jadi",
	},
	"放款成功时间": LangItem{
		LangEnUS: "Payment Time",
		LangIdID: "waktu dana cair",
	},
	"贷款还清时间": LangItem{
		LangEnUS: "Finish Time",
		LangIdID: "waktu dana kembalian",
	},
	"还款": LangItem{
		LangEnUS: "Repayment",
		LangIdID: "kembali dana",
	},
	"还款管理": LangItem{
		LangEnUS: "Repayment Management",
		LangIdID: "kembali dana",
	},
	"还款状态": LangItem{
		LangEnUS: "Repayment Status",
		LangIdID: "status kembalian",
	},
	"账户": LangItem{
		LangEnUS: "Account",
		LangIdID: "rekning",
	},
	"VA账户": LangItem{
		LangEnUS: "VA",
		LangIdID: "rekning",
	},
	"应还日期": LangItem{
		LangEnUS: "Due Date",
		LangIdID: "tanggal harus kembali",
	},
	"应还总额": LangItem{
		LangEnUS: "Total return",
		LangIdID: "jumlah wajib kembali",
	},
	"实还总额": LangItem{
		LangEnUS: "Total amount",
		LangIdID: "jumlah dikembali",
	},
	"减免总额": LangItem{
		LangEnUS: "Total reduction",
		LangIdID: "jumlah potongan",
	},
	"资金流水": LangItem{
		LangEnUS: "Capital Flow",
		LangIdID: "mutasi",
	},
	"还款计划": LangItem{
		LangEnUS: "Repayment Plan",
		LangIdID: "rencana pengembalian",
	},
	"逾期管理": LangItem{
		LangEnUS: "Overdue Management",
		LangIdID: "pengelola keterlambatan",
	},
	"入催": LangItem{
		LangEnUS: "Join Urge",
		LangIdID: "masuk koleksi",
	},
	"出催": LangItem{
		LangEnUS: "Out Urge",
		LangIdID: "keluar koleksi",
	},
	"入催时间": LangItem{
		LangEnUS: "Start urge time",
		LangIdID: "tanggal masuk koleksi",
	},
	"逾期天数": LangItem{
		LangEnUS: "Overdue Days",
		LangIdID: "jangka waktu terlambat",
	},
	"出催起止日期": LangItem{
		LangEnUS: "join urge time",
		LangIdID: "Tanggal berhenti peringatan pengembalian mulai dan akhir",
	},
	"入催起止日期": LangItem{
		LangEnUS: "out urge time",
		LangIdID: "Tanggal peringatan pengembalian mulai dan akhir",
	},
	"实还金额": LangItem{
		LangEnUS: "real repay amount",
		LangIdID: "Dana telah dikembalikan",
	},
	"最后一次催收时间": LangItem{
		LangEnUS: "latest urge time",
		LangIdID: "Waktu peringatan pengembalian terakhir",
	},
	"案件级别": LangItem{
		LangEnUS: "Case Level",
		LangIdID: "tingkatan kasus",
	},
	"催收人": LangItem{
		LangEnUS: "Urge staff",
		LangIdID: "staf pedorong koleksi",
	},
	"实催人": LangItem{
		LangEnUS: "Actually Urge",
		LangIdID: "pedorong koleksi asli",
	},
	"催收情况": LangItem{
		LangEnUS: "Urge record",
		LangIdID: "Koleksi",
	},

	"催收类型": LangItem{
		LangEnUS: "Urge Type",
		LangIdID: "koleksi Ketik",
	},
	"自催": LangItem{
		LangEnUS: "Urge self",
		LangIdID: "koleksi sendiri",
	},
	"委外": LangItem{
		LangEnUS: "Outsourceing",
		LangIdID: "Outsourceing",
	},
	"委外公司": LangItem{
		LangEnUS: "Outsourceing Company",
		LangIdID: "Outsourceing Perusahaan",
	},
	"委外状态": LangItem{
		LangEnUS: "Outsourceing Status",
		LangIdID: "Outsourceing Negara",
	},
	"委外中": LangItem{
		LangEnUS: "Outsourceing ing",
		LangIdID: "Outsourceing ing",
	},
	"已委外": LangItem{
		LangEnUS: "Outsourceing ed",
		LangIdID: "Outsourceing ed",
	},
	"审批状态": LangItem{
		LangEnUS: "Review Status",
		LangIdID: "Review Negara",
	},
	"委外申请起止日期": LangItem{
		LangEnUS: "Outsourceing Apply String-end date",
		LangIdID: "Outsourceing Tanggal mulai dan akhir aplikasi",
	},

	"创建工单": LangItem{
		LangEnUS: "Create Ticket",
		LangIdID: "Buat Target",
	},
	"查看工单": LangItem{
		LangEnUS: "View Ticket",
		LangIdID: "Lihat Target",
	},

	"同意": LangItem{
		LangEnUS: "Agree",
		LangIdID: "Setuju",
	},

	"短信": LangItem{
		LangEnUS: "SMS",
		LangIdID: "sms",
	},
	"发送短信": LangItem{
		LangEnUS: "Send SMS",
		LangIdID: "kirim sms",
	},
	"短信内容": LangItem{
		LangEnUS: "SMS Content",
		LangIdID: "isi sms",
	},
	"借款状态": LangItem{
		LangEnUS: "Loan status",
		LangIdID: "status pinjaman",
	},
	"合同金额": LangItem{
		LangEnUS: "Contract amount",
		LangIdID: "nominal kontrak pinjaman",
	},
	"放款时间": LangItem{
		LangEnUS: "Loan time",
		LangIdID: "waktu dana cairwaktu dana cair",
	},
	"结清时间": LangItem{
		LangEnUS: "Closing time",
		LangIdID: "waktu tutup kasus",
	},
	"修改银行卡号": LangItem{
		LangEnUS: "Modify bank card number",
		LangIdID: "ubah rekning bank",
	},
	"重新放款": LangItem{
		LangEnUS: "Loan again",
		LangIdID: "ulangi cairan",
	},
	"订单": LangItem{
		LangEnUS: "Order",
		LangIdID: "Pesanan",
	},
	"订单类型": LangItem{
		LangEnUS: "Order Type",
		LangIdID: "jenis pesan",
	},
	// 订单状态 {{{
	"已提交申请": LangItem{
		LangEnUS: "Applied",
		LangIdID: "sudah pengajukan",
	},
	"已提交审核": LangItem{
		LangEnUS: "Waiting for approval",
		LangIdID: "sudah varifikasi",
	},
	"审核拒绝": LangItem{
		LangEnUS: "Reject",
		LangIdID: "varifikasi ditolak",
	},
	"等待人工审核": LangItem{
		LangEnUS: "Wait Manual",
		LangIdID: "tunggu varifikasi manual",
	},
	"等待放款": LangItem{
		LangEnUS: "Waiting for payment",
		LangIdID: "tunggu pinjaman cair",
	},
	"放款失败": LangItem{
		LangEnUS: "Payment failure",
		LangIdID: "gagal cair dana",
	},
	"等待还款": LangItem{
		LangEnUS: "Waiting for repayment",
		LangIdID: "tunggu kembalian",
	},
	"已结清": LangItem{
		LangEnUS: "Cleared End",
		LangIdID: "sudah lunas",
	},
	"逾期": LangItem{
		LangEnUS: "OverDue",
		LangIdID: "terlambat",
	},
	"失效": LangItem{
		LangEnUS: "Disable",
		LangIdID: "Disable",
	},
	"部分还款": LangItem{
		LangEnUS: "Partial repayment",
		LangIdID: "kembalian sebagian",
	},
	"正在放款中": LangItem{
		LangEnUS: "Loan is doing",
		LangIdID: "dana sedang cair",
	},
	"放款成功": LangItem{
		LangEnUS: "Payment success",
		LangIdID: "berhasil",
	},
	// }}}
	"业务流水": LangItem{
		LangEnUS: "Business journal",
		LangIdID: "NO. referensi",
	},
	"类型": LangItem{
		LangEnUS: "Type",
		LangIdID: "tipe",
	},
	"其它": LangItem{
		LangEnUS: "other",
		LangIdID: "lain",
	},
	"操作成功": LangItem{
		LangEnUS: "Successful Operation",
		LangIdID: "operasi berhasil",
	},
	"本窗口将在": LangItem{
		LangEnUS: "After",
		LangIdID: "halaman ini akan ditutup otomatis dalam",
	},
	"秒后自动跳转": LangItem{
		LangEnUS: "second, jump auto.",
		LangIdID: "kedua, lompat otomatis",
	},
	"立即跳转": LangItem{
		LangEnUS: "Jump now",
		LangIdID: "Lompat sekarang",
	},

	"秒后自动关闭": LangItem{
		LangEnUS: "second, window will close.",
	},
	"更新时间": LangItem{
		LangEnUS: "update time",
		LangIdID: "waktu perbarui",
	},
	"创建时间": LangItem{
		LangEnUS: "Create time",
		LangIdID: "Waktu pembuatann",
	},
	// 表单显示相关 {{{
	"更新": LangItem{
		LangEnUS: "Update",
		//LangIdID: "halaman ini akan ditutup otomatis dalam",
	},
	"删除": LangItem{
		LangEnUS: "Delete",
		LangIdID: "hapus",
	},
	"搜索": LangItem{
		LangEnUS: "Search",
		LangIdID: "cari",
	},
	"清除": LangItem{
		LangEnUS: "Clear",
		LangIdID: "hapus",
	},
	"移除": LangItem{
		LangEnUS: "move",
		LangIdID: "hapus",
	},
	"未知": LangItem{
		LangEnUS: "Unkown",
		LangIdID: "belum ketahuan",
	},
	"提交": LangItem{
		LangEnUS: "Submit",
		LangIdID: "masuk",
	},
	"取消": LangItem{
		LangIdID: "batal",
	},
	"重置": LangItem{
		LangEnUS: "Reset",
		LangIdID: "reset",
	},
	"刷新": LangItem{
		LangEnUS: "Refresh",
		LangIdID: "perbarui",
	},
	"请选择": LangItem{
		LangEnUS: "Please Choose",
		LangIdID: "silahkan pilih",
	},
	// }}}
	"减免": LangItem{
		LangEnUS: "Relief",
		LangIdID: "kurangi",
	},
	"减免罚息": LangItem{
		LangEnUS: "Relief penalty interest",
		LangIdID: "kurangi bunga denda",
	},
	"减免利息": LangItem{
		LangEnUS: "Relief Interest",
		LangIdID: "kurangi bunga",
	},
	"随机值": LangItem{
		LangEnUS: "Random Value",
		LangIdID: "Random Value",
	},
	"无提示": LangItem{
		LangEnUS: "No hint",
		LangIdID: "tidak ada peritahuan",
	},
	"正常": LangItem{
		LangEnUS: "Normal",
		LangIdID: "normal",
	},
	"异常": LangItem{
		LangEnUS: "Abnormal",
		LangIdID: "aneh",
	},
	// 客户详情
	"客户详情": LangItem{
		LangEnUS: "Customer Detail Information",
		LangIdID: "informasi lengkap nasabah",
	},
	"身份信息": LangItem{
		LangEnUS: "Identity Information",
		LangIdID: "informasi identitas",
	},
	"其他信息": LangItem{
		LangEnUS: "Other Information",
		LangIdID: "informasi lainnya",
	},
	"大数据信息": LangItem{
		LangEnUS: "Big Data Information",
		LangIdID: "informasi big data",
	},
	"沟通记录": LangItem{
		LangEnUS: "Communication record",
		LangIdID: "catatan komunikasi",
	},
	"查重": LangItem{
		LangEnUS: "Check duplicate order",
		LangIdID: "Cek duplikat",
	},
	"性别": LangItem{
		LangEnUS: "Gender",
		LangIdID: "Gender",
	},
	"身份检查结果": LangItem{
		LangEnUS: "Identity check result",
		LangIdID: "hasil periksa identitas",
	},
	"手持照片比对": LangItem{
		LangEnUS: "Handheld photo comparison",
		LangIdID: "perbandingan foto pegang KTP",
	},
	"图像资料": LangItem{
		LangEnUS: "Image data",
		LangIdID: "data image",
	},
	"手持身份证照片": LangItem{
		LangEnUS: "Handheld photo",
		LangIdID: "foto pegang KTP",
	},
	"复贷手持身份证照片": LangItem{
		LangEnUS: "Reloan Handheld photo",
		LangIdID: "pinjaman ulang foto pegang KTP",
	},
	"活体认证最佳": LangItem{
		LangEnUS: "Live certification best",
		LangIdID: "varifikasi aktif terbaik",
	},
	"活体认证全景": LangItem{
		LangEnUS: "Live certification panorama",
		LangIdID: "panorama varifikasi aktif",
	},
	"活体认证参考": LangItem{
		LangEnUS: "Live Certification Reference",
		LangIdID: "referensi varifikasi aktif",
	},
	"工作信息": LangItem{
		LangEnUS: "Work Information",
		LangIdID: "informasi kerja",
	},
	"工作类型": LangItem{
		LangEnUS: "Work Type",
		LangIdID: "sifat kerja",
	},
	"月收入": LangItem{
		LangEnUS: "Monthly Income",
		LangIdID: "gaji bulanan",
	},
	"公司名称": LangItem{
		LangEnUS: "Company Name",
		LangIdID: "nama PT.",
	},
	"公司城市": LangItem{
		LangEnUS: "Company City",
		LangIdID: "kota domisili PT.",
	},
	"公司详细地址": LangItem{
		LangEnUS: "Company Address",
		LangIdID: "alamat lengkap PT. ",
	},
	"工作年限": LangItem{
		LangEnUS: "Working years",
		LangIdID: "pengalaman kerja",
	},
	"联系人信息": LangItem{
		LangEnUS: "Contact Information",
		LangIdID: "informasi kontak",
	},
	"姓名": LangItem{
		LangEnUS: "Name",
		LangIdID: "nama",
	},
	"联系人": LangItem{
		LangEnUS: "Contact",
		LangIdID: "kontak",
	},
	"电话": LangItem{
		LangEnUS: "Phone",
		LangIdID: "telepon",
	},
	"关系": LangItem{
		LangEnUS: "Relationship",
		LangIdID: "relasi",
	},
	"附加信息": LangItem{
		LangEnUS: "Extra Information",
		LangIdID: "informasi tambahan",
	},
	"教育状况": LangItem{
		LangEnUS: "Education status",
		LangIdID: "status pendidikan",
	},
	"婚姻状况": LangItem{
		LangEnUS: "Marital status",
		LangIdID: "status pernikahan",
	},
	"居住城市": LangItem{
		LangEnUS: "Living City",
		LangIdID: "kota domisili",
	},
	"居住详细地址": LangItem{
		LangEnUS: "Residence detailed address",
		LangIdID: "alamat lengkap domisili",
	},
	"子女数": LangItem{
		LangEnUS: "Number of children",
		LangIdID: "jumlah anak",
	},
	// 大数据据
	"未抓取到设备信息": LangItem{
		LangEnUS: "No device information was crawled",
		LangIdID: "gagal dapat data Perangkat",
	},
	"未抓取到通话记录": LangItem{
		LangEnUS: "No call history was crawled",
		LangIdID: "gagal dapat data riwayat komunikasi",
	},
	"未抓取到通讯录": LangItem{
		LangEnUS: "No contacts was crawled",
		LangIdID: "gagal dapat data kontak",
	},
	"未抓取到短信记录": LangItem{
		LangEnUS: "No SMS history was crawled",
		LangIdID: "gagal dapat data sms",
	},
	"未抓取到GPS信息": LangItem{
		LangEnUS: "Uncaught GPS information",
		LangIdID: "gagal dapat data GPS.",
	},
	"近1个月设备号登录的手机账号": LangItem{
		LangEnUS: "Mobile account number registered in the device number for the last month",
		LangIdID: "Perangkat ini masuk berapa nomor hp dalam jangka waktu satu bulan",
	},
	"近1个月手机号登录的设备号": LangItem{
		LangEnUS: "Device number registered in mobile phone number in the last month",
		LangIdID: "Nomor perangkat terdaftar di nomor ponsel dalam sebulan terakhir",
	},
	"3个月内00:00—4:00通话时长": LangItem{
		LangEnUS: "Call duration from 00:00 to 4:00 within 3 months",
		LangIdID: "Durasi panggilan dari 00:00 sampai 4:00 dalam waktu 3 bulan",
	},
	"3个月内无通话记录天数": LangItem{
		LangEnUS: "No call record days in 3 months",
		LangIdID: "harian yang tidak ada rekam panggilan dalam 3 bulan",
	},
	"3个月内与第一联系人通话次数": LangItem{
		LangEnUS: "Number of calls with the first contact within 3 months",
		LangIdID: "Jumlah panggilan dengan kontak paling banyak dihubungi dalam 3 bulan",
	},
	"3个月内与第一联系人通话时长": LangItem{
		LangEnUS: "Talking to the first contact within 3 months",
		LangIdID: "kelamaan berbicara dengan kontakpaling banyak dihubungi  dalam 3 bulan",
	},
	"通讯录个数": LangItem{
		LangEnUS: "Number of address books",
		LangIdID: "Jumlah kontak",
	},
	"通讯录个数小于配置": LangItem{
		LangEnUS: "Number of address books is less than the configuration",
		LangIdID: "kontak simpan kurang dari konfigurasi",
	},
	"通讯录中固定电话占比": LangItem{
		LangEnUS: "The proportion of fixed telephones in address book",
		LangIdID: "Proporsi telepon tetap di buku kontak",
	},
	"通话记录名单占通讯录的比例": LangItem{
		LangEnUS: "List of call records as a percentage of contacts",
		LangIdID: "Daftar catatan panggilan sebagai persentase kontak",
	},
	"包含以下关键字“逾期”的短信数量": LangItem{
		LangEnUS: `The number of text messages with the keyword "overdue" below`,
		LangIdID: `Jumlah pesan dengan kata kunci "terlambat"`,
	},
	"没有抓到大数据": LangItem{
		LangEnUS: "No big data was crawled",
		LangIdID: "tidak ada bigdata",
	},
	"男": LangItem{
		LangEnUS: "male",
		LangIdID: "laki-laki",
	},
	"女": LangItem{
		LangEnUS: "female",
		LangIdID: "perempuan",
	},
	"总计": LangItem{
		LangEnUS: "Total record",
		LangIdID: "jumlah",
	},
	"总页数": LangItem{
		LangEnUS: "Total pages",
		LangIdID: "jumlah halaman",
	},
	"未定义": LangItem{
		LangEnUS: "Undefined",
		LangIdID: "Undefined",
	},
	"案件级别调整": LangItem{
		LangEnUS: "Case Level Adjust",
		LangIdID: "adaptasi tingkatan kasus",
	},
	"催收": LangItem{
		LangEnUS: "Urge",
		LangIdID: "koleksi",
	},
	"催收结果": LangItem{
		LangEnUS: "Urge Result",
		LangIdID: "hasil koleksi",
	},
	"出催原因": LangItem{
		LangEnUS: "Out Urge Reason",
		LangIdID: "alasan koleksi",
	},
	"出催时间": LangItem{
		LangEnUS: "Out Urge Date",
		LangIdID: "waktu koleksi",
	},
	"权限不足": LangItem{
		LangEnUS: "Access Denied",
		LangIdID: "tidak punya akses",
	},
	"筛选": LangItem{
		LangEnUS: "Filter",
		LangIdID: "pilih",
	},
	"全部": LangItem{
		LangEnUS: "All",
		LangIdID: "semua",
	},
	"测试客户": LangItem{
		LangEnUS: "Tester Customer",
		LangIdID: "nasabah percobaan",
	},

	// rbac {
	// 角色权限管理
	"角色权限管理": LangItem{
		LangEnUS: "role rights management",
		LangIdID: "pengelola hak admin masing2",
	},
	"角色未分配权限": LangItem{
		LangEnUS: "role unassigned permissions",
		LangIdID: "akun yang belum diberikan fungsi",
	},
	"角色已分配权限": LangItem{
		LangEnUS: "role assigned permissions",
		LangIdID: "akun yang sudah diberikan fungsi",
	},
	"角色权限分配": LangItem{
		LangEnUS: "role permissions assignment",
		LangIdID: "berbagi fungsi admin",
	},
	// } end

	// 短信管理
	"关联ID": LangItem{
		LangEnUS: "association ID",
		LangIdID: "ID asosiasi",
	},
	"短信服务商": LangItem{
		LangEnUS: "SMS Provider",
		LangIdID: "operator",
	},
	"短信类型": LangItem{
		LangEnUS: "SMS Type",
		LangIdID: "Tipe SMS",
	},
	"发送时间": LangItem{
		LangEnUS: "Send Time",
		LangIdID: "Waktu kirim",
	},
	"发送状态": LangItem{
		LangEnUS: "Send Status",
		LangIdID: "Status kirim",
	},
	"关联": LangItem{
		LangEnUS: "Related",
		LangIdID: "terkait",
	},
	"关联 ID": LangItem{
		LangEnUS: "Related ID",
		LangIdID: "ID terkait",
	},
	"验证码": LangItem{
		LangEnUS: "verifycode",
		LangIdID: "kode verifikasi",
	},
	"验证码管理": LangItem{
		LangEnUS: "verifycode management",
		LangIdID: "Pengelola kode verifikasi",
	},
	"短信发送状态跟踪": LangItem{
		LangEnUS: "SMS Status",
		LangIdID: "Pelacakan status SMS",
	},
	"过期时间": LangItem{
		LangEnUS: "expiration",
		LangIdID: "Waktu kedaluwarsa",
	},

	"登录/注册": LangItem{
		LangEnUS: "login/register",
		LangIdID: "login/pendaftar",
	},
	"请求登录/注册": LangItem{
		LangEnUS: "login/register",
		LangIdID: "masuk/pendaftar",
	},
	"创建订单": {
		LangEnUS: "create order",
		LangIdID: "Buat pesan baru",
	},
	"首贷": {
		LangEnUS: "First loan",
		LangIdID: "Pinjaman pertama",
	},
	"复贷": {
		LangEnUS: "Reloan",
		LangIdID: "Pinjaman ulang",
	},
	"注销": {
		LangEnUS: "logout",
		LangIdID: "Hapus",
	},
	"放贷成功": {
		LangEnUS: "loan success",
		LangIdID: "berhasil cair dana",
	},
	"注册": {
		LangEnUS: "register",
		LangIdID: "Pendaftaran",
	},
	"登录": {
		LangEnUS: "login",
		LangIdID: "login",
	},
	"找回密码": {
		LangEnUS: "find password",
		LangIdID: "lupa kata sandi",
	},
	"短信提醒还款": {
		LangEnUS: "repayment alerted by SMS",
		LangIdID: "Sms ingatan pengembalian",
	},
	"催收短信提醒": {
		LangEnUS: "Collection alerted by SMS",
		LangIdID: "Sms peringatan pengembalian",
	},
	"确认订单": {
		LangEnUS: "confirm order",
		LangIdID: "Konfirmasi pesanan",
	},
	"已验证": {
		LangEnUS: "verified",
		LangIdID: "sudah verifikasi",
	},
	"未验证": {
		LangEnUS: "unverified",
		LangIdID: "belum verifikasi",
	},
	"发送失败": {
		LangEnUS: "send failed",
		LangIdID: "Gagal kirim",
	},
	"验证失败": {
		LangEnUS: "verify failed",
		LangIdID: "Gagal verifikasi",
	},
	"已送达": {
		LangEnUS: "arrived",
		LangIdID: "terkirim",
	},
	"未送达": {
		LangEnUS: "not delivered",
		LangIdID: "Belum terkirim",
	},

	"修改": LangItem{
		LangEnUS: "edit",
		LangIdID: "ubah",
	},
	"修改密码": LangItem{
		LangEnUS: "Change Password",
		LangIdID: "Ubah kata sandi",
	},
	"请输入初始化密码": LangItem{
		LangEnUS: "Please enter your initial password",
	},
	"请输入大于8位密码": LangItem{
		LangEnUS: "Please enter a password greater than 8 bits",
	},
	"请输入数字和字母组合": LangItem{
		LangEnUS: "Please enter a combination of Numbers and letters",
	},
	"请输入确认密码": LangItem{
		LangEnUS: "SMS Status",
	},
	"与初始化密码一样，请重新输入": LangItem{
		LangEnUS: "Please enter the confirmation password",
	},
	"确认密码不正确": LangItem{
		LangEnUS: "Confirm that the password is incorrect",
	},
	"请输入密码": LangItem{
		LangEnUS: "Please enter password",
	},
	"确认修改密码？": LangItem{
		LangEnUS: "Confirm password change?",
	},
	"借款历史": LangItem{
		LangEnUS: "Borrowing history",
		LangIdID: "riwayat pinjaman",
	},
	"电话拨打时间": LangItem{
		LangEnUS: "call time",
		LangIdID: "waktu telephone",
	},
	"电话拨打对象": LangItem{
		LangEnUS: "Target contact",
		LangIdID: "dengan siapa telephone",
	},
	"拨打对象": LangItem{
		LangEnUS: "Target contact",
		LangIdID: "Siapa yang dihubungi ",
	},
	"电话接通情况": LangItem{
		LangEnUS: "phone connection situation",
		LangIdID: "koneksi",
	},
	"是否拨通": LangItem{
		LangEnUS: "is/not connect",
		LangIdID: "Apakah akan menelepon",
	},
	"偿还意愿": LangItem{
		LangIdID: "Kesediaan untuk membayar kembali",
	},
	"接通": LangItem{
		LangEnUS: "connected",
		LangIdID: "terhubung",
	},
	"已接通": LangItem{
		LangEnUS: "connected",
		LangIdID: "terhubung",
	},
	"未接通": LangItem{
		LangEnUS: "unconnected",
		LangIdID: "tidak terhubung",
	},
	"本人": LangItem{
		LangEnUS: "self",
		LangIdID: "pribadi",
	},
	"其他": LangItem{
		LangEnUS: "other",
		LangIdID: "lain-lain",
	},
	"还款意愿": LangItem{
		LangEnUS: "repay inclination",
		LangIdID: "kemauan pengembalian",
	},
	"承诺还款时间": LangItem{
		LangEnUS: "promise repay time",
		LangIdID: "Waktu pembayaran yang dijanjikan",
	},

	"无人接听": LangItem{
		LangEnUS: "Continue Ringing",
		LangIdID: "Terus berdering",
	},
	"逾期原因": LangItem{
		LangEnUS: "overdue reason",
		LangIdID: "alasan lewat jatuh tempo",
	},

	"未接通原因": LangItem{
		LangEnUS: "unconnected reason",
		LangIdID: "alasan tidak bisa di hubungi",
	},
	"黑名单验证": LangItem{
		LangEnUS: "Balcklist",
	},

	"用户修改密码": LangItem{
		LangEnUS: "change password",
		LangIdID: "ubah kata sandi akun",
	},
	"原密码": LangItem{
		LangEnUS: "original password",
		LangIdID: "kata sandi lama",
	},
	"新密码": LangItem{
		LangEnUS: "new password",
		LangIdID: "Kata sandi baru",
	},
	"确认密码": LangItem{
		LangEnUS: "confirm password",
		LangIdID: "ulangi kata sandi",
	},
	"系统管理": LangItem{
		LangEnUS: "system management",
		LangIdID: "pengelola sistem admin",
	},
	"用户列表": LangItem{
		LangEnUS: "user list",
		LangIdID: "daftar administrasi",
	},
	"后台用户管理": LangItem{
		LangEnUS: "user management",
		LangIdID: "pengelola akun administrasi",
	},
	"印尼电核人员": LangItem{
		LangEnUS: "Indonesia auditors",
		LangIdID: "admin verifikasi Indonesia",
	},
	"风控人员": LangItem{
		LangEnUS: "risk controller",
		LangIdID: "Admin manajer resiko",
	},
	"无效": LangItem{
		LangEnUS: "invalid",
		LangIdID: "tidak berkalu",
	},
	"封禁": LangItem{
		LangEnUS: "forbidden",
		LangIdID: "blokir",
	},
	"解封": LangItem{
		LangEnUS: "unblock",
		LangIdID: "buka",
	},
	"更新后台用户": LangItem{
		LangEnUS: "update blackground user",
		LangIdID: "perbarui akun admin",
	},
	"邮箱": LangItem{
		LangEnUS: "email",
		LangIdID: "email",
	},
	"昵称": LangItem{
		LangEnUS: "nickname",
		LangIdID: "nama",
	},
	"角色": LangItem{
		LangEnUS: "role",
		LangIdID: "fungsi",
	},
	"选择一个角色": LangItem{
		LangEnUS: "choice role",
		LangIdID: "pilih salah satu fungsi",
	},
	"监管": LangItem{
		LangEnUS: "supervise",
		LangIdID: "pengawas",
	},
	"风控": LangItem{
		LangEnUS: "risk control",
		LangIdID: "resiko manajer",
	},
	"请设置一个复杂密码": LangItem{
		LangEnUS: "set complex password",
		LangIdID: "buatkan password kompleks",
	},
	"不更新密码直接留空": LangItem{
		LangEnUS: "not update password and leave it blank",
		LangIdID: "kosongkan untuk tidak diperbarui ",
	},
	"监管准备": LangItem{
		LangEnUS: "regulatory preparation",
		LangIdID: "persiapan pengawas",
	},
	"印尼催收人员": LangItem{
		LangEnUS: "Indonesian collection personnel",
		LangIdID: "koleksi",
	},
	"风控管理人员": LangItem{
		LangEnUS: "Risk management staff",
		LangIdID: "resiko manajer",
	},
	"客服主管": LangItem{
		LangEnUS: "customer service supervisor",
		LangIdID: "supervisor layanan pelanggan",
	},
	"风控主管": LangItem{
		LangEnUS: "risk control supervisor",
		LangIdID: "pengawas kontrol risiko",
	},
	"运营主管": LangItem{
		LangEnUS: "director of operations",
		LangIdID: "manajer operasional",
	},
	"联系方式": LangItem{
		LangEnUS: "phone",
		LangIdID: "kontak ponsel",
	},
	"添加后台用户": LangItem{
		LangEnUS: "add user",
		LangIdID: "perbarui akun admin",
	},
	"添加用户": LangItem{
		LangEnUS: "add user",
		LangIdID: "perbarui akun admin",
	},
	"登陆密码": LangItem{
		LangEnUS: "login password",
		LangIdID: "Kata sandi login",
	},
	"第三方黑名单通过": LangItem{
		LangEnUS: "third-party blacklist passed",
		LangIdID: "tidak diblokir pihak ketiga",
	},
	"第三方黑名单拒绝": LangItem{
		LangEnUS: "third-party blacklist rejected",
		LangIdID: "keblokir oleh pihak ketiga",
	},
	"第三方黑名单处理中": LangItem{
		LangEnUS: "third-party blacklist processing",
		LangIdID: "sedang proses blacklist pihak ketiga",
	},
	"等待第三方黑名单验证": LangItem{
		LangEnUS: "waiting for verifing by third-party blacklist",
		LangIdID: "Menunggu verifikasi daftar hitam pihak ketiga",
	},
	"有": LangItem{
		LangEnUS: "exist",
		LangIdID: "ada",
	},
	"无": LangItem{
		LangEnUS: "no",
		LangIdID: "tidak ada",
	},
	"无人接通": LangItem{
		LangEnUS: "no answer",
		LangIdID: "Tidak ada jawaban",
	},
	"应还宽限期利息": LangItem{
		LangEnUS: "grace period interest should be repaid",
		LangIdID: "bunga masa grace yang wajib kembalian",
	},
	"已还宽限期利息": LangItem{
		LangEnUS: "grace period interest already repaid",
		LangIdID: "bunga masa grace yang sudah dibalik",
	},
	"已减本金": LangItem{
		LangEnUS: "reduced principal",
		LangIdID: "tunggakan sudah dipotong",
	},
	"已减宽限期利息": LangItem{
		LangEnUS: "reduced grace period interest",
		LangIdID: "bunga masa grace yang sudah dipotong",
	},
	"已减罚息": LangItem{
		LangEnUS: "reduced penalty interest",
		LangIdID: "bunga denda sudah dipotong",
	},
	"入账": LangItem{
		LangEnUS: "account",
		LangIdID: "kredit",
	},
	"出账": LangItem{
		LangEnUS: "debit",
		LangIdID: "Debit",
	},
	"摩比神奇": LangItem{
		LangEnUS: "Mobi",
		LangIdID: "Mobi",
	},
	"利息": LangItem{
		LangEnUS: "interest",
		LangIdID: "Bunga",
	},
	"省份": LangItem{
		LangEnUS: "province",
		LangIdID: "provinsi",
	},
	"城市": LangItem{
		LangEnUS: "city",
		LangIdID: "kota",
	},
	"区": LangItem{
		LangEnUS: "area",
		LangIdID: "kabupaten",
	},
	"村": LangItem{
		LangEnUS: "village",
		LangIdID: "desa",
	},
	"减免本金": LangItem{
		LangEnUS: "reducing principal",
		LangIdID: "sudah verifikasi",
	},

	"直接挂断": LangItem{
		LangEnUS: "hang up directly",
		LangIdID: "Tidak diangkat",
	},

	"占线": LangItem{
		LangEnUS: "busy",
		LangIdID: "SIBUK",
	},

	"返回": LangItem{
		LangEnUS: "go back",
		LangIdID: "Kembali",
	},

	"实还本金": LangItem{
		LangEnUS: "principal should be repaid",
		LangIdID: "tunggakan sudah dibalik",
	},

	"实还宽限期利息": LangItem{
		LangEnUS: "grace period interest already repaid",
		LangIdID: "Masa tenggang biaya layanan juga telah dibayarkan",
	},

	"实还罚息": LangItem{
		LangEnUS: "penalty interest already repaid",
		LangIdID: "bunga denda sudah dibalik",
	},

	"减免日期": LangItem{
		LangEnUS: "relief date",
		LangIdID: "Tanggal bantuan",
	},

	"已减免本金": LangItem{
		LangEnUS: "reduced principal",
		LangIdID: "Mengurangi pokok",
	},

	"已减免宽限期利息": LangItem{
		LangEnUS: "grace period interest reduced",
		LangIdID: "Biaya layanan grace period remisi",
	},

	"已减免罚息": LangItem{
		LangEnUS: "penalty interest reduced",
		LangIdID: "Pengecualian dari hukuman",
	},

	"剩余可减免本金": LangItem{
		LangEnUS: "remaining principal can be reduced",
		LangIdID: "Sisa pokok yang dapat dikurang",
	},

	"剩余可减免宽限期利息": LangItem{
		LangEnUS: "remaining grace period interest can be reduced",
		LangIdID: "Sisanya bisa dikurangi Bunga",
	},

	"剩余可减免罚息": LangItem{
		LangEnUS: "remaining penalty interest can be reduced",
		LangIdID: "Sisanya bisa dikurangi Penalti bunga",
	},

	"减免宽限期利息": LangItem{
		LangEnUS: "relief for grace period interest",
		LangIdID: "kurangi bunga",
	},

	// 系统配置
	"系统配置": LangItem{
		LangEnUS: "system configuration",
		LangIdID: "sistem Konfigurasi",
	},
	"配置名称": LangItem{
		LangEnUS: "configuration name",
		LangIdID: "nama Konfigurasi",
	},
	"生效状态": LangItem{
		LangEnUS: "effective status",
		LangIdID: "status efektif",
	},
	"新增配置": LangItem{
		LangEnUS: "add configuration",
		LangIdID: "Konfigurasi tambahan",
	},
	"配置项名称": LangItem{
		LangEnUS: "configuration item name",
		LangIdID: "nama item Konfigurasi",
	},
	"配置项描述": LangItem{
		LangEnUS: "configuration item description",
		LangIdID: "diskrepsi item Konfigurasi",
	},
	"配置项值类型": LangItem{
		LangEnUS: "configuration item type",
		LangIdID: "jenis item Konfigurasi",
	},
	"配置项值": LangItem{
		LangEnUS: "configuration item value",
		LangIdID: "nilai item Konfigurasi",
	},
	"显示权重": LangItem{
		LangEnUS: "weight",
		LangIdID: "berat",
	},
	"后台显示权重": LangItem{
		LangEnUS: "weight",
		LangIdID: "berat",
	},
	"上线时间": LangItem{
		LangEnUS: "online time",
		LangIdID: "waktu efektif",
	},
	"下线时间": LangItem{
		LangEnUS: "offline time",
		LangIdID: "waktu cabut",
	},
	"有效状态": LangItem{
		LangEnUS: "effective status",
		LangIdID: "status berlaku",
	},
	"合法变量名": LangItem{
		LangEnUS: "legal variable name",
		LangIdID: "nama variabel hukum",
	},
	"配置项简要描述": LangItem{
		LangEnUS: "brief description of configuration item",
		LangIdID: "deskripsi singkat tentang item konfigurasi",
	},
	"收起": LangItem{
		LangEnUS: "collapse",
		LangIdID: "ciutkan",
	},
	"随机放款的概率,百分值": LangItem{
		LangEnUS: "probability of random loan, percentage value",
		LangIdID: "uang acak Probabilitas, persentase",
	},
	"首贷客户允许的最大借款金额": LangItem{
		LangEnUS: "max loan amount of first loan customers",
		LangIdID: "nominal diizinkan pd pinjaman pertama kali",
	},
	"首贷客户允许的最长借款期限": LangItem{
		LangEnUS: "max loan period of first loan customers",
		LangIdID: "tenor maksimal diizinkan pd pinjaman pertama kali",
	},
	"电核问题级别2级": LangItem{
		LangEnUS: "Phone verify question level 2",
		LangIdID: "pertanyaan verifikasi tingkat 2",
	},
	"不提供服务的区域,城市名称": LangItem{
		LangEnUS: "unserviced area, city name",
		LangIdID: "daerah yang tidak dilayan, nama kota",
	},
	"日订单熔断值": LangItem{
		LangEnUS: "quantity limit of day loan orders",
		LangIdID: "Batas kuantitas pesanan harian",
	},
	"一级随机数命中概率": LangItem{
		LangEnUS: "hit probability of first-level random numbers",
		LangIdID: "Probabilitas klik nomor acak tingkat 1",
	},
	"二级随机数命中概率": LangItem{
		LangEnUS: "hit probability of second-level random numbers",
		LangIdID: "Probabilitas klik nomor acak tingkat 2",
	},
	"预期天数>=n，自动加入黑名单": LangItem{
		LangEnUS: "overdue days ≥ N, join the blacklist automatically",
		LangIdID: "hari keterlambatan lewat≥N, langsuang masuk ke daftar keblokiran",
	},
	"是否启用第三方黑名单": LangItem{
		LangEnUS: "enable third party blacklist",
		LangIdID: "apakah gunakan daftar keblokiran pihak ketiga",
	},
	"第三方黑名单命中范围": LangItem{
		LangEnUS: "hit range of third party blacklist",
		LangIdID: "jangkauwan klik daftar keblokiran pihak ketiga",
	},
	"跳过电核随机数概率": LangItem{
		LangEnUS: "probability of skipping electric random number",
		LangIdID: "Probabilitas lewat verifikasi acak",
	},
	"首贷 反欺诈规则列表,不可为空,以逗号分割;": LangItem{
		LangEnUS: "risk control regular list of first loan, not empty, separated by commas",
		LangIdID: "standar anti-penipuan pada pertama pinjaman, wajib isi, dipisah dengan koma",
	},
	"复贷-随机-有逾期[用户]反欺诈规则列表, 无逾期时为空,直接跳过审查": LangItem{
		LangEnUS: "reloan/test/overdue[user]risk control regular list is empty when no overdue, skip review directly",
		LangIdID: "daftar ajukan lagi/ percobaan/ keterlambatan[nasabah] anti-penipuan, dikosongan pada yang belum terlambat, langsung skip verifikasi",
	},
	"复贷-非随机[用户]反欺诈规则列表, 不走D009": LangItem{
		LangEnUS: "reloan/test/overdue[user]risk control regular list, not execute D009",
		LangIdID: "daftar ajukan lagi/ percobaan/ keterlambatan[nasabah] anti-penipuan, tidak masuk D009",
	},
	"sms 发送策略配置": LangItem{
		LangEnUS: "policy configuration of sending SMS",
		LangIdID: "Konfigurasi Strategi kiriman SMS",
	},
	"反欺诈规则级别1级": LangItem{
		LangEnUS: "anti-fraud regular level 1",
		LangIdID: "standar Anti-penipuan tingkatan 1",
	},
	"反欺诈规则级别2级": LangItem{
		LangEnUS: "anti-fraud regular level 2",
		LangIdID: "standar Anti-penipuan tingkatan 2",
	},
	"风控反欺诈规则级别2级": LangItem{
		LangEnUS: "anti-fraud regular of risk control level 2",
		LangIdID: "standar Anti-penipuan risiko manajemen tingkatan 2",
	},
	"包含以下关键字“逾期”“贷款”的短信数量": LangItem{
		LangEnUS: "anti-fraud regular of risk control level 2",
		LangIdID: "jumlah SMS berisi kata kunci 'terlambat', 'pinjam'",
	},
	"1小时内，账户移动距离≥配置": LangItem{
		LangEnUS: "moving distance of account within one hour is greater than or equal to the configuration",
		LangIdID: "dalam 1 jam, jarak pindah nasabah ≥ konfigurasi",
	},
	"1小时内，设备移动距离≥配置": LangItem{
		LangEnUS: "moving distance of device within one hour is greater than or equal to the configuration",
		LangIdID: "dalam 1 jam, jarak pindah nasabah ≥ konfigurasi",
	},
	"1天内，同一设备注册时间间隔<配置项秒数": LangItem{
		LangEnUS: "registered interval of one device within one day is less than the configuration",
		LangIdID: "dalam 1 hari, jarak waktu buat akun dengan perangkat sama < detik item dikonfigurasi",
	},
	"1天内，同一设备注册的账号数≥配置数": LangItem{
		LangEnUS: "registered nums of one device within one day is greater than or equal to the configuration",
		LangIdID: "dalam 1 hari, jumlah akun dibuat dengan perangkat sama ≥ jumlah dikonfigurasi",
	},
	"1天内，同一设备登录账户号≥配置值": LangItem{
		LangEnUS: "login accounts of one device within one day is greater than or equal to the configuration",
		LangIdID: "dalam 1 hari, jumlah akun dimasuk ke perangkat sama ≥ nilai dikonfigurasi",
	},
	"历史同一设备号登录账号≥配置值": LangItem{
		LangEnUS: "login historical accounts of one device is greater than or equal to the configuration",
		LangIdID: "riwayat akun yang pernah masuk di 1 perangkat ≥ nilai dikonfigurasi",
	},
	"1天内，同一账号登录的设备数≥配置值": LangItem{
		LangEnUS: "login devices of one account within one day is greater than or equal to the configuration",
		LangIdID: "dalam 1 hari, akun sama masuk ke perangkat berbeda ≥ nilai dikonfigurasi",
	},
	"历史同一账号登录的设备数≥配置值": LangItem{
		LangEnUS: "login historical devices of one account is greater than or equal to the configuration",
		LangIdID: "riwayat akun sama masuk ke jumlah perangkat berbeda ≥ nilai dikonfigurasi",
	},
	"同一IP，1小时内注册设备数≥配置值": LangItem{
		LangEnUS: "registered device num of one IP within one hour is greater than or equal to the configuration",
		LangIdID: "IP sama, perangkat yg daftar dalam 1jam  ≥ nilai dikonfigurasi",
	},

	"同一IP，1小时内注册账号数≥配置值": LangItem{
		LangEnUS: "registered account num of one IP within one hour is greater than or equal to the configuration",
		LangIdID: "IP sama, akun yg dibuat dalam 1jam  ≥ nilai dikonfigurasi",
	},
	"3个月内呼入与呼出前10的重叠个数≤配置值": LangItem{
		LangEnUS: "Overlapping nums of incoming and outgoing calls which are top 10 within 3 months is less than or equal to the configuration",
		LangIdID: "dalam 3bulan， 10 terbanyak panggil dan terima nomor yang sama ≤ nilai dikonfigurasi",
	},
	"7天内，同一设备注册的账号数≥配置值": LangItem{
		LangEnUS: "registered account num of one device within 7 days is greater than or equal to the configuration",
		LangIdID: "dalam 7 hari, jumlah akun dibuat dengan perangkat sama ≥ jumlah dikonfigurasi",
	},
	"30天内，同一设备注册的账号数≥配置值": LangItem{
		LangEnUS: "registered account num of one device within 30 days is greater than or equal to the configuration",
		LangIdID: "dalam 30 hari, jumlah akun dibuat dengan perangkat sama ≥ jumlah dikonfigurasi",
	},

	"7天内，同一设备登录账户号≥配置值": LangItem{
		LangEnUS: "login account num of one device within 7 days is greater than or equal to the configuration",
		LangIdID: "dalam 7 hari, jumlah akun dimasuk ke perangkat sama ≥ nilai dikonfigurasi",
	},
	"30天内，同一设备登录账户号≥配置值": LangItem{
		LangEnUS: "login account num of one device within 30 days is greater than or equal to the configuration",
		LangIdID: "dalam 30 hari, jumlah akun dimasuk ke perangkat sama ≥ nilai dikonfigurasi",
	},
	"最近通话时间距现在天数>n": LangItem{
		LangEnUS: "the days that last call from now is greater than configuration value 'n'",
		LangIdID: "jarak harian yang terakhir telepon sampai sekarang > N",
	},
	"3个月内无短信记录天数>=n": LangItem{
		LangEnUS: "the days that no message record within 3 months is greater than or equal to the configuration value 'n'",
		LangIdID: "harian yang komunikasi via sms dalam 3bulan ≥ N",
	},
	"最近发短信时间距离现在天数>n": LangItem{
		LangEnUS: "the days that last send message from now is greater than configuration value 'n'",
		LangIdID: "jarak waktu terakhir kirim sms sampai sekarang > N",
	},

	"近1个月内同一单位的申请人数限制": LangItem{
		LangEnUS: "max reply num of one company within one month",
		LangIdID: "dalam waktu 1 bulan, batasan peminjam dalam satu perusahaan",
	},
	"近1个月内同一居住地址的申请人数": LangItem{
		LangEnUS: "max reply num of one residence within one month",
		LangIdID: "dalam waktu 1 bulan, jumlah peminjam dalam alamat yang sama",
	},
	"近1个月同联系人在我司申请人数≥配置值": LangItem{
		LangEnUS: "reply num of the same contract in my company within 1 month is greater than or equal to the configuration",
		LangIdID: "dalam waktu 1 bulan, jumlah peminjam dengan PIC yang sama ≥ konfigurasi",
	},
	"近3个月同联系人在我司申请人数≥配置值": LangItem{
		LangEnUS: "reply num of the same contract in my company within 3 months is greater than or equal to the configuration",
		LangIdID: "dalam waktu 3 bulan, jumlah peminjam dengan PIC yang sama ≥ konfigurasi",
	},

	"近3个月内同一居住地址的申请人数≥配置值": LangItem{
		LangEnUS: "reply num of the same residence with 3 months is greater than or equal to the configuration",
		LangIdID: "dalam waktu 3 bulan, jumlah peminjam dalam alamat yang sama ≥ konfigurasi",
	},
	"历史同一居住地址的申请人数≥配置值": LangItem{
		LangEnUS: "reply historical num of one residence is greater than or equal to the configuration",
		LangIdID: "riwayat jumlah peminjam dalam alamat yang sama ≥ konfigurasi",
	},
	"近3月内同单位名称在我司申请人数≥配置值": LangItem{
		LangEnUS: "reply num of the same company within 3 months is greater than or equal to the configuratin",
		LangIdID: "dalam waktu 3 bulan, batasan peminjam dalam satu perusahaan ≥ konfigurasi",
	},
	"第一联系人在我司贷款历史最大逾期天数≥配置值": LangItem{
		LangEnUS: "max overdue days of the first contract in my company is greater than or equal to the configuration",
		LangIdID: "hari keterlambatan maksimal dari pinjaman PIC pertama ≥ nilai konfigurasi",
	},

	"第二联系人在我司贷款历史最高逾期天数≥配置值": LangItem{
		LangEnUS: "max overdue days of the second contract in my company is greater than or equal to the configuration",
		LangIdID: "hari keterlambatan maksimal dari pinjaman PIC kedua ≥ nilai konfigurasi",
	},
	"同联系人我司申请人当前逾期人数≥配置值": LangItem{
		LangEnUS: "overdue num of the same contract  is greater than or equal to the configuration",
		LangIdID: "jumlah keterlambatan peminjam dengan PIC yang sama  ≥ nilai konfigurasi",
	},
	"同居住地址的申请人当前逾期人数≥配置值": LangItem{
		LangEnUS: "overdue num of one residence is greater than or equal to the configuration",
		LangIdID: "jumlah keterlambatan peminjam dengan alamat tinggal yang sama  ≥ nilai konfigurasi",
	},
	"同单位名称我司申请人当前逾期人数≥配置值": LangItem{
		LangEnUS: "overdue num of the same company is greater than or equal to the configuratin",
		LangIdID: "jumlah keterlambatan peminjam dengan nama perusahaan yang sama  ≥ nilai konfigurasi",
	},
	"三个月内累计逾期订单数≥配置值": LangItem{
		LangEnUS: "overdue order num within 3 months is greater than or equal to the configuration",
		LangIdID: "jumlah pesan keterlambatan dalam 3 bulan  ≥ nilai konfigurasi",
	},
	"同一银行账号关联客户数>=2": LangItem{
		LangEnUS: "customer nums associated with the same bank account is greater than or equal to 2",
		LangIdID: "rekning yang sama terkait nasabah ≥ 2",
	},
	"系统评分不足 为 A卡评分<600": LangItem{
		LangEnUS: "system score is insufficient which A card score is less than 600",
		LangIdID: "scoring sistem kurang dari nilai kartu A < 600",
	},

	// 权限组管理
	"权限组管理": LangItem{
		LangEnUS: "rights group management",
		LangIdID: "pengelola klompok hak",
	},
	"权限组名": LangItem{
		LangEnUS: "rights group name",
		LangIdID: "nama klompok hak",
	},
	"新增权限组": LangItem{
		LangEnUS: "add rights group",
		LangIdID: "tambah klompok hak",
	},
	"特殊权限": LangItem{
		LangEnUS: "special permissions",
		LangIdID: "hak khusus",
	},
	"短信管理": LangItem{
		LangEnUS: "sms management",
		LangIdID: "pengelola sms",
	},
	"菜单管理": LangItem{
		LangEnUS: "menu management",
		LangIdID: "pengelola menu",
	},
	"名称": LangItem{
		LangEnUS: "name",
		LangIdID: "nama",
	},

	//逾期有条件减免
	"结清减免": LangItem{
		LangEnUS: "Reduced when repay complete",
		LangIdID: "Potongan pelunasan",
	},
	"应还罚息和宽限期利息": LangItem{
		LangEnUS: "Repay penalty and grace period interest",
		LangIdID: "Denda dan bunga masa toleransi",
	},
	"可减免金额": LangItem{
		LangEnUS: "Can be reduced amount",
		LangIdID: "Jumlah dana yang dapat dipotong",
	},
	"结清最低应还款项": LangItem{
		LangEnUS: "Repay complete lowst amount",
		LangIdID: "Jumlah dana pelunasan minimal",
	},
	"减免配额": LangItem{
		LangEnUS: "Reduced quota",
		LangIdID: "Jumlah kesempatan pengajuan potongan",
	},
	"结清减免申请成功": LangItem{
		LangEnUS: "Repay complete apply success",
		LangIdID: "Pengajuan potongan pelunasan disetujui.",
	},
	"今日可减免的客户数已达上限,如有需要,请与主管申请": LangItem{
		LangEnUS: "Reduced quota limited, if you need, apply again to your leader",
		LangIdID: "Jumlah Nasabah yang dapat potongan telah sampai batas hari ini, kalau perlu, mohon ajukan kepada pengurus.",
	},
	"已申请结清减免": LangItem{
		LangEnUS: "Already apply repay complete reduced",
		LangIdID: "Telah mengajukan potongan pelunasan",
	},
	"该案件已申请结清减免": LangItem{
		LangEnUS: "This case aleardy apply the repay complete reduced.",
		LangIdID: "Pesanan ini telah mengajukan potongan pelunasan.",
	},
	"该案件不允许申请结清减免": LangItem{
		LangEnUS: "This case can't apply the repay complete reduced.",
		LangIdID: "Pesanan ini tidak mengizinkan pengiriman diskon pembayaran.",
	},

	// 操作日志
	"操作日志": LangItem{
		LangEnUS: "operation log",
		LangIdID: "catatan harian operasi",
	},
	"日志ID": LangItem{
		LangEnUS: "log ID",
		LangIdID: "nomor catatan",
	},
	"表名": LangItem{
		LangEnUS: "table name",
		LangIdID: "nama daftar",
	},
	"操作用户ID": LangItem{
		LangEnUS: "operator ID",
		LangIdID: "admin ID",
	},

	// 工单
	"工单": LangItem{
		LangEnUS: "Ticket",
		LangIdID: "Target",
	},
	"开始": LangItem{
		LangEnUS: "Start",
		LangIdID: "Mulai",
	},
	"完成": LangItem{
		LangEnUS: "Complete",
		LangIdID: "Selesai",
	},
	"工单管理": LangItem{
		LangEnUS: "Ticket Manage",
		LangIdID: "Target Manage",
	},
	"工单分类": LangItem{
		LangEnUS: "Ticket Category",
		LangIdID: "Kategori Target",
	},
	"分配给": LangItem{
		LangEnUS: "Assign To",
		LangIdID: "bagi ke",
	},
	"相关ID": LangItem{
		LangEnUS: "Related ID",
		LangIdID: "ID terkait",
	},
	"分配时间": LangItem{
		LangEnUS: "Assign Time",
		LangIdID: "Waktu alokasi",
	},
	"开始时间": LangItem{
		LangEnUS: "Start Time",
		LangIdID: "Waktu mulai",
	},
	"完成时间": LangItem{
		LangEnUS: "Complete Time",
		LangIdID: "Waktu penyelesaian",
	},
	"关闭时间": LangItem{
		LangEnUS: "Close Time",
		LangIdID: "Waktu penutupan",
	},
	"关闭原因": LangItem{
		LangEnUS: "Close Reason",
		LangIdID: "Penyebab penutupan",
	},
	"直达链接": LangItem{
		LangEnUS: "Quick Link",
		LangIdID: "link",
	},
	"分配": LangItem{
		LangEnUS: "Assign",
		LangIdID: "Alokasi",
	},
	"关闭": LangItem{
		LangEnUS: "Close",
		LangIdID: "Tertutup",
	},
	"我的工单": LangItem{
		LangEnUS: "My Ticket",
		LangIdID: "Target saya",
	},
	"人员绩效管理": LangItem{
		LangEnUS: "Staff performance",
		LangIdID: "Pengelolaan Kinerja Staf",
	},
	"人员当日绩效": LangItem{
		LangEnUS: "Staff day performance",
		LangIdID: "Kinerja hari staf",
	},
	"我的绩效": LangItem{
		LangEnUS: "My performance",
		LangIdID: "Pengelolaan Kinerja Saya",
	},
	"风险评级": LangItem{
		LangEnUS: "Risk Level",
		LangIdID: "Tingkat risiko",
	},
	"工作状态管理": LangItem{
		LangEnUS: "Work Status Manage",
		LangIdID: "Manage Status Kerja ",
	},
	"工作状态": LangItem{
		LangEnUS: "Work Status",
		LangIdID: "Status Kerja ",
	},
	"已创建": LangItem{
		LangEnUS: "Already Created",
		LangIdID: "Sudah Dibuat",
	},
	"已分配": LangItem{
		LangEnUS: "Already Assigned",
		LangIdID: "Sudah bagi",
	},
	"进行中": LangItem{
		LangEnUS: "Porccessing",
		LangIdID: "sedang proses",
	},
	"已完成": LangItem{
		LangEnUS: "Already Completed",
		LangIdID: "Sudah Selesai",
	},
	"已关闭": LangItem{
		LangEnUS: "Already Closed",
		LangIdID: "Sudah Ditutup",
	},
	"分配工单": LangItem{
		LangEnUS: "Assign Ticket",
		LangIdID: "Bagi Target",
	},
	"如订单ID,客户ID等": LangItem{
		LangEnUS: "As Order ID,Customer ID ...",
		LangIdID: "Seperti ID Pesan , ID Pelanggan dll. ",
	},
	"暂停工作": LangItem{
		LangEnUS: "Stop Work",
		LangIdID: "Berhenti bekerja",
	},
	"恢复正常": LangItem{
		LangEnUS: "Return to work",
		LangIdID: "Kembali bekerja",
	},
	"批量分配": LangItem{
		LangEnUS: "Batch Assign",
		LangIdID: "",
	},
	"选择": LangItem{
		LangEnUS: "Select",
		LangIdID: "",
	},
	"已选中": LangItem{
		LangEnUS: "Already selected ",
		LangIdID: "",
	},
	"是否在线": LangItem{
		LangEnUS: "Is Online",
		LangIdID: "",
	},
	"在线": LangItem{
		LangEnUS: "Online",
		LangIdID: "Online",
	},
	"全选": LangItem{
		LangEnUS: "All",
		LangIdID: "",
	},
	"还款计划历史": LangItem{
		LangEnUS: "Repay History",
		LangIdID: "Riwayat rencana pembayaran kembali",
	},
	"序号": LangItem{
		LangEnUS: "No.",
		LangIdID: "Tidak",
	},
	"出帐日期": LangItem{
		LangEnUS: "Pay Out Date",
		LangIdID: "Tanggal penagihan",
	},
	"入账日期": LangItem{
		LangEnUS: "Pay In Date",
		LangIdID: "Tanggal masuk",
	},
	"实还服务费": LangItem{
		LangEnUS: "Repayment service fee",
		LangIdID: "biaya admin sudah dikembalian",
	},
	"实还利息": LangItem{
		LangEnUS: "Repayment Interest",
		LangIdID: "bunga yang sudah dibalik",
	},
	// ticket new
	"异常工单": LangItem{
		LangEnUS: "Abnormal Ticket",
		LangIdID: "Order abnormal",
	},
	"无需处理": LangItem{
		LangEnUS: "No need handle",
		LangIdID: "Tidak usah ditangani",
	},
	"案件升级": LangItem{
		LangEnUS: "Case Up",
		LangIdID: "Perkara meningkat",
	},
	"恢复接单": LangItem{
		LangEnUS: "Continue Accept Ticket",
		LangIdID: "Terima order kembali",
	},
	"暂停接单": LangItem{
		LangEnUS: "Stop Accept Ticket",
		LangIdID: "Berhenti terima order sementara",
	},
	"是": LangItem{
		LangEnUS: "Yes",
		LangIdID: "Ya",
	},
	"否": LangItem{
		LangEnUS: "No",
		LangIdID: "Tidak",
	},
	"上次登录时间": LangItem{
		LangEnUS: "Last Login Time",
		LangIdID: "Waktu login terakhir",
	},
	"接收工单状态": LangItem{
		LangEnUS: "Accept Ticket Status",
		LangIdID: "Accept Ticket Status",
	},
	"交流方式": LangItem{
		LangEnUS: "Communication Way",
		LangIdID: "Cara hubungi",
	},
	"是否是空号": LangItem{
		LangEnUS: "Is Empty Number",
		LangIdID: "Nomor terdaftar atau tidak",
	},
	"处理次数": LangItem{
		LangEnUS: "Handle Times",
		LangIdID: "Jumlah Kali Konmunikasi",
	},
	"上次处理时间": LangItem{
		LangEnUS: "Last Handle Time",
		LangIdID: "Terakhir komunikasi ",
	},
	"下次沟通时间": LangItem{
		LangEnUS: "Next Communication Time",
		LangIdID: "Komunikasi Berikut",
	},
	"客户期望电核时间": LangItem{
		LangEnUS: "Best Recall Time",
		LangIdID: "Waktu Penarikan Terbaik",
	},
	"部分完成时间": LangItem{
		LangEnUS: "Partial Complete Time",
		LangIdID: "Waktu pembayaran bagian",
	},
	"部分完成": LangItem{
		LangEnUS: "Partial Complete",
		LangIdID: "Bagian selesai",
	},
	"分案本金": LangItem{
		LangEnUS: "Total Left Unpaid Principal",
		LangIdID: "Dana target penagihan",
	},
	"回款本金": LangItem{
		LangEnUS: "Repay Principal",
		LangIdID: "Modal yang telah dikembalikan",
	},
	"回款息费": LangItem{
		LangEnUS: "Repay Interest",
		LangIdID: "Bunga yang telah dikembalikan",
	},
	"回款总金额": LangItem{
		LangEnUS: "Total Repay Amount",
		LangIdID: "Jumlah dana yang telah dikembalikan",
	},
	"回收率": LangItem{
		LangEnUS: "Repay Rate",
		LangIdID: "Rasio pengembalian penagihan",
	},
	"目标回收率": LangItem{
		LangEnUS: "Target Repay Rate",
		LangIdID: "Rasio target pengembalian penagihan",
	},
	"差值金额": LangItem{
		LangEnUS: "Diff Target Repay Amount",
		LangIdID: "Dana belum dikembalikan dari target",
	},
	"总分配量": LangItem{
		LangEnUS: "Total Assign Num",
		LangIdID: "Total distribusi",
	},
	"处理数": LangItem{
		LangEnUS: "Handle Num",
		LangIdID: "Jumlah penanganan",
	},
	"完成数": LangItem{
		LangEnUS: "Complete Num",
		LangIdID: "Jumlah penyelesaian",
	},
	"负载数": LangItem{
		LangEnUS: "Load Num",
		LangIdID: "Jumlah target ditugaskan",
	},
	"新分配数": LangItem{
		LangEnUS: "New Assign Num",
		LangIdID: "Jumlah distribusi baru",
	},
	"总处理量": LangItem{
		LangEnUS: "Total Handle Num",
		LangIdID: "Total penanganan",
	},
	"个人整体情况": LangItem{
		LangEnUS: "Overview",
		LangIdID: "Keadaan petugas",
	},
	"员工日进度表": LangItem{
		LangEnUS: "Daily Process",
		LangIdID: "Proses harian",
	},
	"回收进度": LangItem{
		LangEnUS: "Repay Process",
		LangIdID: "Proses pengembalian penagihan",
	},
	"时间/Hour": LangItem{
		LangEnUS: "Time/Hour",
		LangIdID: "Waktu/Jam",
	},
	"工作进度": LangItem{
		LangEnUS: "Work Process",
		LangIdID: "Kemajuan kerja",
	},
	"今日到手Bonus": LangItem{
		LangEnUS: "Daily Bonus",
		LangIdID: "Bonus hari ini",
	},
	"排名": LangItem{
		LangEnUS: "Ranking",
		LangIdID: "Peringkat",
	},
	"当前排名": LangItem{
		LangEnUS: "Ranking",
		LangIdID: " Peringkat sekarang",
	},
	"催收工单": LangItem{
		LangEnUS: "Urge Ticket",
		LangIdID: "Mendesak Tiket",
	},
	"催收次数": LangItem{
		LangEnUS: "Urge number",
		LangIdID: "Jumlah frekuensi penagihan",
	},
	"上次催收时间": LangItem{
		LangEnUS: "Last urge time",
		LangIdID: "Waktu penagihan terakhir",
	},
	"下一次沟通时间": LangItem{
		LangEnUS: "Next call time",
		LangIdID: "Waktu pembayaran yang dijanjikan",
	},
	"上次催收记录": LangItem{
		LangEnUS: "Last urge record",
		LangIdID: "Catatan penagihan terakhir",
	},

	"客户类型": LangItem{
		LangEnUS: "Customer type",
		LangIdID: "Jenis pelanggan",
	},
	"新工单": LangItem{
		LangEnUS: "NEW",
		LangIdID: "BARU",
	},
	"旧工单": LangItem{
		LangEnUS: "OLD",
		LangIdID: "TUA",
	},
	"申请展期": LangItem{
		LangEnUS: "Apply extension",
		LangIdID: "Ajukan perpanjang waktu",
	},

	"已完成 ": LangItem{
		LangEnUS: "COMPLETE",
		LangIdID: "LENGKAP",
	},
	"日期": LangItem{
		LangEnUS: "Date",
		LangIdID: "Tanggal",
	},
	"小组情况": LangItem{
		LangEnUS: "Team Progress",
		LangIdID: "Kemajuan tim",
	},
	"小组回收率": LangItem{
		LangEnUS: "Team Repay Rate",
		LangIdID: "Tingkat pemulihan grup",
	},
	"小组名称": LangItem{
		LangEnUS: "Team Name",
		LangIdID: "Nama grup",
	},
	"小组成员排名": LangItem{
		LangEnUS: "Team Member Ranking",
		LangIdID: "Peringkat grup",
	},

	// 还款提醒
	"剩余应还总额": LangItem{
		LangEnUS: "Remaining repayment amount",
		LangIdID: "Jumlah dana harus dilunasi",
	},
	"提醒结果": LangItem{
		LangEnUS: "Remind Result",
		LangIdID: "Balasan peringatan",
	},
	"提醒": LangItem{
		LangEnUS: "Remind",
		LangIdID: "Peringatan",
	},
	"提醒记录": LangItem{
		LangEnUS: "Remind Result",
		LangIdID: "Sejarah peringatan",
	},
	"自动外呼记录": LangItem{
		LangEnUS: "Auto Call Record",
		LangIdID: "Rekam Panggilan Otomatis",
	},
	"自动外呼时间": LangItem{
		LangEnUS: "Auto Call Time",
		LangIdID: "Waktu Panggilan Otomatis",
	},
	"催收方式": LangItem{
		LangEnUS: "Communication tools",
		LangIdID: "Alat komunikasi",
	},
	"未还原因": LangItem{
		LangEnUS: "none repay reason",
		LangIdID: "Alasan tidak membayar",
	},
	"回调延迟": LangItem{
		LangEnUS: "Callback delay",
		LangIdID: "delay",
	},
	"客户未收到放款": LangItem{
		LangEnUS: "Customer did't receive the money",
		LangIdID: "Belum menerima Dana",
	},

	"上次提醒记录": LangItem{
		LangEnUS: "Last remind record",
		LangIdID: "Daftar Peringatan Terakhir",
	},
	"拨打状态": LangItem{
		LangEnUS: "Calling status",
		LangIdID: "Status Panggilan",
	},
	"提醒工单": LangItem{
		LangEnUS: "Remind Ticket",
		LangIdID: "Daftar Kerja Mengingatkan Nasabah",
	},
	"电核&信息查看工单": LangItem{
		LangEnUS: "PV&InfoReview Ticket",
		LangIdID: "PV&InfoReview Tiket",
	},
	"检查图片": LangItem{
		LangEnUS: "Check photo",
		LangIdID: "Periksa foto",
	},

	//第三方
	"第三方调用列表": LangItem{
		LangEnUS: "Thirdparty Request List",
		LangIdID: "Daftar panggilan pihak ketiga",
	},
	"第三方请求详情": LangItem{
		LangEnUS: "Thirdparty Request Detail",
		LangIdID: "Detail permintaan pihak ketiga",
	},
	"第三方": LangItem{
		LangEnUS: "Thirdparty",
		LangIdID: "Pihak ketiga",
	},
	"关系ID": LangItem{
		LangEnUS: "Related ID",
		LangIdID: "ID Hubungan",
	},
	"请求": LangItem{
		LangEnUS: "Request",
		LangIdID: "Permintaan",
	},
	"响应体": LangItem{
		LangEnUS: "Response",
		LangIdID: "Tanggapi",
	},

	// 用户反馈
	"用户反馈": LangItem{
		LangEnUS: "Customer Feedback",
		LangIdID: "komentar pelanggan",
	},
	"订单申请时间": LangItem{
		LangEnUS: "Order Apply Time",
		LangIdID: "Waktu pendaftaran pinjaman",
	},
	"申请次数": LangItem{
		LangEnUS: "Apply Num",
		LangIdID: "Frekuensi pinjaman",
	},
	"申请成功次数": LangItem{
		LangEnUS: "Apply Success Num",
		LangIdID: "Pinjaman yang berhasil",
	},
	"反馈列表": LangItem{
		LangEnUS: "Feedback List",
		LangIdID: "Daftar komentar",
	},
	"反馈管理": LangItem{
		LangEnUS: "Feedback Manage",
		LangIdID: "pengelola komentar",
	},
	"反馈分类": LangItem{
		LangEnUS: "Feedback Category",
		LangIdID: "kategori komentar",
	},
	"内容": LangItem{
		LangEnUS: "Content",
		LangIdID: "isi",
	},
	"APP版本": LangItem{
		LangEnUS: "App Version",
		LangIdID: "Versi App",
	},
	"API版本": LangItem{
		LangEnUS: "API Version",
		LangIdID: "Versi API",
	},
	"导出": LangItem{
		LangEnUS: "Export",
		LangIdID: "Export",
	},
	"借款订单": LangItem{
		LangEnUS: "Borrowing  Order",
		LangIdID: "Borrowing  Order",
	},
	"临时订单": LangItem{
		LangEnUS: "Provisional Order",
		LangIdID: "Provisional Order",
	},
	"剩余还款金额小于等于": LangItem{
		LangEnUS: "The remaining amount less equal than",
		LangIdID: "Sisa pembayaran kurang dan sama dengan",
	},
	"展期结清": LangItem{
		LangEnUS: "Rolling Clear",
		LangIdID: "Pelunasan pada masa tunda",
	},
	"冻结": LangItem{
		LangEnUS: "Frozen",
		LangIdID: "Bekukan",
	},
	"展期申请中": LangItem{
		LangEnUS: "Rolling Apply",
		LangIdID: "Permintaan masa tunda sedang diproses",
	},
	"展期失效": LangItem{
		LangEnUS: "Rolling Disable",
		LangIdID: "Permintaan masa tunda dibatalkan otoritas",
	},
	"等待展期": LangItem{
		LangEnUS: "Rolling",
		LangIdID: "Penungguan masa tunda",
	},
	"国家": LangItem{
		LangEnUS: "Country",
		LangIdID: "Negara",
	},
	"ID检索": LangItem{
		LangEnUS: "ID to retrieve",
		LangIdID: "Penelusuran ID",
	},
	"文本检索": LangItem{
		LangEnUS: "The Text Retrieval",
		LangIdID: "Penelusuran teks",
	},
	"字符数": LangItem{
		LangEnUS: "Number Of Characters",
		LangIdID: "Jumlah karakter",
	},
	"每页条数": LangItem{
		LangEnUS: "Number Each Page",
		LangIdID: "Jumlah baris setiap halaman",
	},
	"用户反馈内容": LangItem{
		LangEnUS: "User Feedback",
		LangIdID: "Isi umpan balik",
	},
	"是否允许展期": LangItem{
		LangEnUS: "Rolling Allow Or Not",
		LangIdID: "Apakah akan mengizinkan ekstensi",
	},
	"允许展期": LangItem{
		LangEnUS: "Rolling Allow",
		LangIdID: "Izinkan ekstensi",
	},
	"不允许展期": LangItem{
		LangEnUS: "Rolling Not Allow",
		LangIdID: "Tidak ada ekstensi yang diizinkan",
	},

	"拒接": LangItem{
		LangEnUS: "Reject",
		LangIdID: "Ditolak",
	},
	"用户不存在": LangItem{
		LangEnUS: "Not Registered",
		LangIdID: "Belum Terdaftar",
	},
	"客户设置拒接所有来电": LangItem{
		LangEnUS: "Customer Block All Incoming Call",
		LangIdID: "Pelanggan menutup seluruh panggilan masuk",
	},
	"拨打后返回主页面": LangItem{
		LangEnUS: "Back to main screen",
		LangIdID: "Kembali ke Menu Utama",
	},
	"有还款意愿": LangItem{
		LangEnUS: "Willing to pay",
		LangIdID: "Bersedia Membayar",
	},

	"不是客户本人接听": LangItem{
		LangEnUS: "Not the customer",
		LangIdID: "Bukan Pelanggan",
	},
	"接听后挂断": LangItem{
		LangEnUS: "Hang up",
		LangIdID: "Ditutup Seketika",
	},
	"超市付款码": LangItem{
		LangEnUS: "PaymentCode",
		LangIdID: "Nomor Tagihan Alfamart ",
	},
	"付款状态": LangItem{
		LangEnUS: "Pay status",
		LangIdID: "Status Pembayaran",
	},
	"未付款": LangItem{
		LangEnUS: "Pending",
		LangIdID: "Menunggu Pembayaran",
	},
	"已付款": LangItem{
		LangEnUS: "Paid",
		LangIdID: "Pembayaran Berhasil",
	},
	"状态码": LangItem{
		LangEnUS: "Status Code",
		LangIdID: "Kode status",
	},
	"第三方管理": LangItem{
		LangEnUS: "Third party management",
		LangIdID: "Manajemen pihak ketiga",
	},
	"请求参数": LangItem{
		LangEnUS: "Request parameter",
		LangIdID: "Meminta parameter",
	},
	"详情": LangItem{
		LangEnUS: "Details",
		LangIdID: "Detail",
	},
	"客户标签": LangItem{
		LangIdID: "Label pelanggan",
	},
	"修改手机号": LangItem{
		LangEnUS: "Modify mobile",
		LangIdID: "Mengubah nomor ponsel",
	},
	"原手机号": LangItem{
		LangEnUS: "Original mobile",
		LangIdID: "Nomor ponsel lama",
	},
	"新手机号": LangItem{
		LangEnUS: "New mobile",
		LangIdID: "Nomor ponsel baru",
	},
	"确认新手机号": LangItem{
		LangEnUS: "Confirm mobile",
		LangIdID: "Konfirmasi nomor ponsel baru",
	},
	"手机号为空": LangItem{
		LangEnUS: "Mobile empty",
		LangIdID: "Belum memasukkan nomor ponsel",
	},
	"两次手机号不一致": LangItem{
		LangEnUS: "Twice mobile are different",
		LangIdID: "Nomor ponsel kedua berbeda",
	},
	"便利店付款码": LangItem{
		LangEnUS: "Market Payment Code",
		LangIdID: "Nomor Tagihan",
	},
	"工单信息": LangItem{
		LangEnUS: "Ticket Info",
		LangIdID: "Informasi Pesanan",
	},
	"上次承诺还款时间": LangItem{
		LangEnUS: "Last promise reapy time",
		LangIdID: "Waktu Pengembalian Terjanjikan Terakhir",
	},
	"案件评级": LangItem{
		LangEnUS: "Case level",
		LangIdID: "Tingkat kasus",
	},
	"复制": LangItem{
		LangEnUS: "Copy",
		LangIdID: "Salinan",
	},
	"申请委外": LangItem{
		LangEnUS: "Apply Outsource",
		LangIdID: "Terapkan Outsource",
	},

	"记录催收结果": LangItem{
		LangEnUS: "Log Urge Detail",
		LangIdID: "Catatan Hasil Peringatan Pengembalian",
	},
	"还款信息": LangItem{
		LangEnUS: "Repayment Info",
		LangIdID: "Informasi Pengembalian",
	},
	"公司电话": LangItem{
		LangEnUS: "Company Phone Number",
		LangIdID: "Nomor Telepon Perusahaan ",
	},
	"发薪日": LangItem{
		LangEnUS: "Salary day",
		LangIdID: "Hari Penerimaan gaji",
	},
	"保存催收记录": LangItem{
		LangEnUS: "Save Urge Detail",
		LangIdID: "Simpan Catatan Peringatan Pengembalian",
	},
	"选择日期": LangItem{
		LangEnUS: "Select date",
		LangIdID: "Pilih Tanggal",
	},
	"通讯录": LangItem{
		LangEnUS: "Address Book",
		LangIdID: "Buku Alamat",
	},
	"借款信息": LangItem{
		LangEnUS: "Loan Info",
		LangIdID: "Informasi Peminjaman",
	},
	"剩余可处理天数": LangItem{
		LangEnUS: "Left Can Handle Days",
		LangIdID: "Jumlah Hari Bisa Menangani Tersisa",
	},
	"VA账号": LangItem{
		LangEnUS: "VA Account",
		LangIdID: "Nomor VA",
	},
	"付款码失效时间": LangItem{
		LangEnUS: "Payment Code Expire Time",
		LangIdID: "Waktu Nomor Tagihan Kedaluwarsa",
	},
	"减免后待结清金额": LangItem{
		LangEnUS: "Left Repayment Amount After Relief",
		LangIdID: "Tagihan Setelah Pengurangan",
	},
	"展期需还款金额": LangItem{
		LangEnUS: "Roll Need Repayment Amount",
		LangIdID: "Tagihan Untuk Pengajuan Masa Tunda",
	},
	"没有钱": LangItem{
		LangEnUS: "No Money",
		LangIdID: "belum punya uang",
	},
	"小孩子要上学": LangItem{
		LangEnUS: "Children need to go school",
		LangIdID: "karena kebutuhan anak sekolah",
	},
	"没发工资": LangItem{
		LangEnUS: "No salary",
		LangIdID: "belum menerima gaji",
	},
	"去医院(生病、出事故、家里有人生病)": LangItem{
		LangEnUS: "Go Hospital(Ill, Accident on families)",
		LangIdID: "karena sakit,kecelakaan, musibah dalam keluarga",
	},
	"忘记有借款": LangItem{
		LangEnUS: "Forget the loan",
		LangIdID: "karena lupa  dia ada pinjaman di rupiah cepa",
	},
	"不知道应该如何操作还款/ATM尝试还款但失败": LangItem{
		LangEnUS: "No idea to operate repay/ATM trial but failed",
		LangIdID: "Nasabah tidak tahu cara melakukan pembayaran/selalu gagal waktu pembayaran lewat ATM",
	},
	//催收结果
	"承诺还款": LangItem{
		LangEnUS: "Willing to pay",
		LangIdID: "Janji Bayar",
	},
	"无还款意愿": LangItem{
		LangEnUS: "Not Willing to pay",
		LangIdID: "Tak Mau Bayar",
	},
	"非借款人": LangItem{
		LangEnUS: "Not Customer",
		LangIdID: "Bukan pelanggan",
	},
	"振铃不接": LangItem{
		LangEnUS: "Continue Ringing",
		LangIdID: "Berdering",
	},
	"不在服务区": LangItem{
		LangEnUS: "Out Of Coverage Area",
		LangIdID: "Tak Dapat Dihubungi/Di Luar Jangkauan",
	},
	"暂时无法接通": LangItem{
		LangEnUS: "Not Active",
		LangIdID: "Tak Aktif",
	},
	"呼叫转接": LangItem{
		LangEnUS: "Call forwarding",
		LangIdID: "Dialihkan",
	},
	"用户忙": LangItem{
		LangEnUS: "Busy",
		LangIdID: "Sedang Sibuk",
	},
	"被拉黑": LangItem{
		LangEnUS: "Blocked all coming call",
		LangIdID: "Diblok",
	},
	"空号": LangItem{
		LangEnUS: "Not Registered (Invalid)",
		LangIdID: "Belum Terdaftar / Nomor Salah",
	},
	"未收到放款": LangItem{
		LangEnUS: "Not receive the money",
		LangIdID: "Belum Terima Dana",
	},
	"已付款未入系统": LangItem{
		LangEnUS: "Already pay but delay",
		LangIdID: "Sudah lakukan Pembayaran tapi Pending",
	},
	"已留言联系人或通讯录": LangItem{
		LangEnUS: "Left message to adressbook",
		LangIdID: "Tinggal Pesan Sama CP/Adressbook",
	},
	"意愿展期": LangItem{
		LangEnUS: "Extension",
		LangIdID: "Mau Perpanjangan",
	},
	"意愿部分还款": LangItem{
		LangEnUS: "Partial payment first",
		LangIdID: "Mau Bayar Sebagian Dulu ",
	},
	"没声音": LangItem{
		LangEnUS: "Nobody answer",
		LangIdID: "Tidak Ada Suara",
	},

	"展期申请成功": LangItem{
		LangEnUS: "Roll Success",
		LangIdID: "Pengajuan perpanjangan telah berhasil",
	},
	"该时间段不允许展期": LangItem{
		LangEnUS: "Invalid Time",
		LangIdID: "Permintaan dapat diajukan antara pukul 8:00 sampai pukul 24:00.",
	},
	// voip相关
	"呼叫": LangItem{
		LangEnUS: "Call",
		LangIdID: "Memanggil",
	},
	"中止": LangItem{
		LangEnUS: "Stop",
		LangIdID: "Menghentikan",
	},
	"记录人": LangItem{
		LangEnUS: "Recorder",
		LangIdID: "Pencatat",
	},
	"手工": LangItem{
		LangEnUS: "Manul",
		LangIdID: "Manual",
	},
	"联系人通讯方式为空": LangItem{
		LangEnUS: "Contact mobile is blank",
		LangIdID: "Kontak belum diisi",
	},
	"获取工单信息失败": LangItem{
		LangEnUS: "Get ticket info fail",
		LangIdID: "Pengambilan informasi target gagal",
	},
	"员工未分配分机号": LangItem{
		LangEnUS: "Not assign extension",
		LangIdID: "Belum didistribusikan nomor ekstensi",
	},
	"获取分机状态失败": LangItem{
		LangEnUS: "Get sip status fail",
		LangIdID: "Pengambilan status ekstensi gagal",
	},
	"获取账户信息失败": LangItem{
		LangEnUS: "Get account info fail",
		LangIdID: "Pengambilan informasi akun gagal",
	},
	"插入分机通话记录失败": LangItem{
		LangEnUS: "Insert call record fail",
		LangIdID: "Penyisipan riwayat panggilan ekstensi gagal",
	},
	"呼叫请求发送失败": LangItem{
		LangEnUS: "Send call request fail",
		LangIdID: "Pengiriman permintaan memanggil gagal",
	},
	"正在呼叫中": LangItem{
		LangEnUS: "Calling",
		LangIdID: "Sedang memanggil",
	},
	"如挂断电话后未自动获取结果，请点击中止并手动选择结果": LangItem{
		LangEnUS: "If the result is not automatically obtained after hanging up the phone, please click ‘stop’ and manually select the result.",
		LangIdID: "Jika setelah panggilan ditutup tidak dapat hasil secara otomatis,silakan hentikan dan pilih hasilnya secera manual",
	},
	"分机管理": LangItem{
		LangEnUS: "sip management",
		LangIdID: "Pengaturan ekstensi",
	},
	"用户姓名": LangItem{
		LangEnUS: "user name",
		LangIdID: "Nama pengguna",
	},
	"分机号码": LangItem{
		LangEnUS: "extnumber",
		LangIdID: "Nomor ekstensi",
	},
	"分机状态": LangItem{
		LangEnUS: "sip status",
		LangIdID: "Status ekstensi",
	},
	"分机通话状态": LangItem{
		LangEnUS: "sip call status",
		LangIdID: "Status panggilan ekstensi",
	},
	"是否启用": LangItem{
		LangEnUS: "is/not enable",
		LangIdID: "Pengaktifan",
	},
	"启用": LangItem{
		LangEnUS: "enable",
		LangIdID: "Aktifkan",
	},
	"禁用": LangItem{
		LangEnUS: "disable",
		LangIdID: "Nonaktifkan",
	},
	"分配状态": LangItem{
		LangEnUS: "assign status",
		LangIdID: "Status distribusi",
	},
	"取消分配": LangItem{
		LangEnUS: "cancel assign",
		LangIdID: "Membatalkan distribusi",
	},
	"分配历史": LangItem{
		LangEnUS: "assign history",
		LangIdID: "Riwayat distribusi",
	},
	"未分配": LangItem{
		LangEnUS: "unassigned",
		LangIdID: "Belum didistribusikan",
	},
	"分机未注册": LangItem{
		LangEnUS: "not registered",
		LangIdID: "Belum terdaftar",
	},
	"空闲": LangItem{
		LangEnUS: "available",
		LangIdID: "Tersedia",
	},
	"振铃": LangItem{
		LangEnUS: "ringing",
		LangIdID: "Mendering",
	},
	"摘机": LangItem{
		LangEnUS: "off-hook",
		LangIdID: "Diangkat",
	},
	"通话中": LangItem{
		LangEnUS: "busy",
		LangIdID: "Sibuk",
	},
	"⾮通话中不能语⾳评分": LangItem{
		LangEnUS: "other",
		LangIdID: "Tidak dapat menilai tanpa menelepon",
	},
	"队列异常": LangItem{
		LangEnUS: "queue exception",
		LangIdID: "Barisan abnormal",
	},
	"⾮法队列": LangItem{
		LangEnUS: "illegal queue",
		LangIdID: "Barisan ilegal",
	},
	"未接听": LangItem{
		LangEnUS: "not answered",
		LangIdID: "Tidak terhubung",
	},
	"等待中": LangItem{
		LangEnUS: "waiting",
		LangIdID: "Sedang menunggu",
	},
	"接听中": LangItem{
		LangEnUS: "receiving",
		LangIdID: "Sedang diterima",
	},
	"已接听": LangItem{
		LangEnUS: "answered",
		LangIdID: "Telah diangkat",
	},
	"暂停": LangItem{
		LangEnUS: "pause",
		LangIdID: "Berhenti sementara",
	},
	"分机已启用": LangItem{
		LangEnUS: "enable",
		LangIdID: "Sedang aktif",
	},
	"分机未启用": LangItem{
		LangEnUS: "not enable",
		LangIdID: "Belum aktif",
	},
	"分配人员": LangItem{
		LangEnUS: "assign staff",
		LangIdID: "Karyawan distribusi",
	},
	"取消分配时间": LangItem{
		LangEnUS: "unassigned time",
		LangIdID: "Waktu membatalkan distribusi",
	},
	"通话记录查询": LangItem{
		LangEnUS: "call record",
		LangIdID: "Riwayat panggilan",
	},
	"拨打时间": LangItem{
		LangEnUS: "starttime",
		LangIdID: "Waktu pemanggilan",
	},
	"话单ID": LangItem{
		LangEnUS: "bill id",
		LangIdID: "ID pemanggilan",
	},
	"主叫号码": LangItem{
		LangEnUS: "display number",
		LangIdID: "Nomor telepon memanggil",
	},
	"目标号码": LangItem{
		LangEnUS: "dest number",
		LangIdID: "Nomor tujuan",
	},
	"呼叫方向": LangItem{
		LangEnUS: "direction",
		LangIdID: "Tujuan pemanggilan",
	},
	"呼出": LangItem{
		LangEnUS: "call out",
		LangIdID: "Keluar",
	},
	"呼入": LangItem{
		LangEnUS: "call in",
		LangIdID: "Masuk",
	},
	"呼叫方法": LangItem{
		LangEnUS: "call method",
		LangIdID: "Cara pemanggilan",
	},
	"应答时间": LangItem{
		LangEnUS: "answer time",
		LangIdID: "Waktu telepon diangkat",
	},
	"结束时间": LangItem{
		LangEnUS: "end time",
		LangIdID: "Waktu telepon ditutup",
	},
	"是否接通": LangItem{
		LangEnUS: "is/not connect",
		LangIdID: "Terhubung atau tidak",
	},
	"通话时长": LangItem{
		LangEnUS: "call duration",
		LangIdID: "Panjang waktu menelepon",
	},
	"接通前等待时长": LangItem{
		LangEnUS: "waiting duration",
		LangIdID: "Panjang waktu menelepon",
	},
	"挂机方向": LangItem{
		LangEnUS: "hangup direction",
		LangIdID: "Penutup telepon",
	},
	"挂机原因": LangItem{
		LangEnUS: "hangup cause",
		LangIdID: "Sebab menutup",
	},
	"录音文件": LangItem{
		LangEnUS: "record file",
		LangIdID: "Dokumen rekaman",
	},
	"正常挂断": LangItem{
		LangEnUS: "hangup normally",
		LangIdID: "Ditutup normal",
	},
	"呼叫取消": LangItem{
		LangEnUS: "call cancellation",
		LangIdID: "Panggilan dibatalkan",
	},
	"拒绝接听": LangItem{
		LangEnUS: "refuse to answer",
		LangIdID: "Panggilan ditolak",
	},
	"外呼通道线路失败": LangItem{
		LangEnUS: "outbound channel line failed",
		LangIdID: "Saluran keluar gagal",
	},
	"用户超时未接听": LangItem{
		LangEnUS: "user timeout missed",
		LangIdID: "Pengguna tidak terhubung dengan melampaui waktu",
	},
	"服务器端挂断": LangItem{
		LangEnUS: "server hang up",
		LangIdID: "Ditutup server",
	},
	"目标不可达": LangItem{
		LangEnUS: "unreachable target",
		LangIdID: "Tujuan tidak terjangkau",
	},
	"定时器超时": LangItem{
		LangEnUS: "timer timeout",
		LangIdID: "Pencatat waktu melampaui batas",
	},
	"呼入时回调接口错误": LangItem{
		LangEnUS: "callback interface error when calling",
		LangIdID: "Penghubung masuk panggilan salah",
	},
	"分机不存在": LangItem{
		LangEnUS: "extnumber not existed",
		LangIdID: "Pesawat telepon sambungan tidak ada",
	},
	"未发现": LangItem{
		LangEnUS: "not found",
		LangIdID: "Tidak ditemui",
	},
	"请求超时": LangItem{
		LangEnUS: "request timeout",
		LangIdID: "pemanggilan lewat waktu",
	},
	"呼叫失效": LangItem{
		LangEnUS: "call invalidation",
		LangIdID: "Pemanggilan gagal",
	},
	"归属地未知": LangItem{
		LangEnUS: "unknown attribution",
		LangIdID: "Lokasi tidak diketahui",
	},
	"其他原因": LangItem{
		LangEnUS: "other reasons",
		LangIdID: "Sebab lain",
	},
	"错误请求": LangItem{
		LangEnUS: "wrong request",
		LangIdID: "Permintaan salah",
	},
	"呼叫被禁止": LangItem{
		LangEnUS: "call is forbidden",
		LangIdID: "Panggilan dilarang",
	},
	"号码被改变": LangItem{
		LangEnUS: "the number was changed",
		LangIdID: "Nomor telepon diubah",
	},
	"呼叫拦截": LangItem{
		LangEnUS: "call interception",
		LangIdID: "Mencegat panggilan",
	},
	"主叫挂机": LangItem{
		LangEnUS: "caller hangup",
		LangIdID: "Ditutup pemanggil",
	},
	"被叫挂机": LangItem{
		LangEnUS: "user called hangup",
		LangIdID: "Ditutup penerima telepon",
	},
	"其他联系人": LangItem{
		LangEnUS: "other contacts",
		LangIdID: "Penghubung lain",
	},
	"工单未分配": LangItem{
		LangEnUS: "ticket unassined",
		LangIdID: "Perintah kerja tidak ditugaskan",
	},
	"如挂断电话后呼叫按钮不可用，请点击中止，方可重新呼叫": LangItem{
		LangEnUS: "If the call button is not available after the call, click stop to re-call.",
		LangIdID: "Jika setelah menutup telepon dan tombol memanggil tidak bekerja, silakan klik tombol menghentikan dan baru bisa menmanggil kembali",
	},
	"请求参数错误": LangItem{
		LangEnUS: "Request parameter error",
		LangIdID: "Permintaan parameter salah",
	},
	"获取token失败": LangItem{
		LangEnUS: "Failed to get token",
		LangIdID: "Pengambilan token salah",
	},
	"请求发送失败": LangItem{
		LangEnUS: "Request failed to send",
		LangIdID: "Pengiriman permintaan gagal",
	},
	"获取分机信息失败": LangItem{
		LangEnUS: "Failed to get sip information",
		LangIdID: "Pengambilan informasi ekstensi gagal",
	},
	"分机呼叫失败": LangItem{
		LangEnUS: "call failed",
		LangIdID: "Menganggil ekstensi gagal",
	},
	"获取通话详单失败": LangItem{
		LangEnUS: "Failed to get the call record",
		LangIdID: "Pengambilan detail panggilan gagal",
	},
	"获取录音文件下载地址失败": LangItem{
		LangEnUS: "Failed to get the recording file download address",
		LangIdID: "Pengambilan situs mengunduh rekaman suara gagal",
	},
	"分机不可用": LangItem{
		LangEnUS: "extnumber is not available",
		LangIdID: "Ekstensi tidak tersedia",
	},
	"更新分机信息失败": LangItem{
		LangEnUS: "Update sip information failed",
		LangIdID: "Update informasi ekstensi gagal",
	},
	"取消分配成功": LangItem{
		LangEnUS: "Unassignment succeeded",
		LangIdID: "Pembatalan distribusi berhasil",
	},
	"不存在可分配分机的用户": LangItem{
		LangEnUS: "no users able to be assigned",
		LangIdID: "Tidak ada pengguna yang dapat didistribusikan ekstensi",
	},
	"此电话号码不允许做外呼操作，如有疑问，请询问leader": LangItem{
		LangEnUS: "This phone number is not allowed to make outbound calls. If in doubt, please ask the leader",
		LangIdID: "Nomor telepon ini tidak boleh menelepon keluar, jika ada pertanyaan, silakan tanya leader.",
	},
	"减免管理": LangItem{
		LangEnUS: "Reduce Manage",
		LangIdID: "Kelola peringanan",
	},
	"CO2案件管理": LangItem{
		LangEnUS: "CO2 Case Manage",
		LangIdID: "CO2 Kasus peringanan",
	},
	"委外审批": LangItem{
		LangEnUS: "Entrust Approval",
		LangIdID: "Komisi Persetujuan",
	},

	"案件": LangItem{
		LangEnUS: "case",
		LangIdID: "Kasus",
	},
	"减免类型": LangItem{
		LangEnUS: "Reduce type",
		LangIdID: "Jenis peringanan",
	},
	"减免状态": LangItem{
		LangEnUS: "Reduce status",
		LangIdID: "Status peringanan",
	},
	"申请人": LangItem{
		LangEnUS: "Apply Uid",
		LangIdID: "Pengaju",
	},
	"审核人": LangItem{
		LangEnUS: "Confirm Uid",
		LangIdID: "Petugas",
	},
	"生效": LangItem{
		LangEnUS: "Enable",
		LangIdID: "Berlaku",
	},
	"自动": LangItem{
		LangEnUS: "auto",
		LangIdID: "Otomatis",
	},
	"结清": LangItem{
		LangEnUS: "Prereduce",
		LangIdID: "Pelunasan",
	},
	"普通": LangItem{
		LangEnUS: "Normal",
		LangIdID: "Biasa",
	},
	"请选择审批意见": LangItem{
		LangEnUS: "please select option",
		LangIdID: "Silakan pilih komentar persetujuan",
	},
	"减免审核": LangItem{
		LangEnUS: "Reduce Confirm",
		LangIdID: "Ulasan penggantian",
	},
	"审批意见": LangItem{
		LangEnUS: "Option",
		LangIdID: "Pendapat persetujuan",
	},
	"审批说明": LangItem{
		LangEnUS: "Remark",
		LangIdID: "Instruksi persetujuan",
	},
	"通过": LangItem{
		LangEnUS: "Approved",
		LangIdID: "Izinkan",
	},
	"拒绝": LangItem{
		LangEnUS: "Reject",
		LangIdID: "Tolak",
	},
	"退款": LangItem{
		LangEnUS: "Refund",
		LangIdID: "Pengembalian dana",
	},
	"退款类型": LangItem{
		LangEnUS: "Refund type",
		LangIdID: "Jenis pengembalian dana",
	},
	"退款金额": LangItem{
		LangEnUS: "Refund amount",
		LangIdID: "Dana diharap dikembalikan",
	},
	"退款到借款订单": LangItem{
		LangEnUS: "Refund to order",
		LangIdID: "Mengembalikan dana ke target pinjaman",
	},
	"退款到银行账户": LangItem{
		LangEnUS: "Refund to bank code",
		LangIdID: "Mengembalikan dana ke rekening bank",
	},
	"金额错误": LangItem{
		LangEnUS: "Amount err",
		LangIdID: "Jumlah dana salah",
	},
	"请选择退款凭证": LangItem{
		LangEnUS: "Please select a refund voucher ",
		LangIdID: "Silakan unggah bukti",
	},
	"扣除手续费": LangItem{
		LangEnUS: "Fee",
		LangIdID: "Dikurangi biaya pelayanan",
	},
	"多还款凭证": LangItem{
		LangEnUS: "Multiple repayment certificate",
		LangIdID: "Bukti pembayaran kelebihan",
	},
	"名字包含数字,不合法": LangItem{
		LangEnUS: "The name contains a number and is illegal",
		LangIdID: " Nama mengandung angka, tidak valid",
	},
	"验证码类型": LangItem{
		LangEnUS: "Verify code type",
		LangIdID: "Jenis kode verifikasi",
	},
	"语音验证码": LangItem{
		LangEnUS: "voice verify code",
		LangIdID: "Kode verifikasi suara",
	},
	"短信验证码": LangItem{
		LangEnUS: "Sms verify code",
		LangIdID: "Kode verifikasi SMS",
	},
	"INVALID_DESTINATION": LangItem{
		LangEnUS: "INVALID_DESTINATION",
		LangIdID: "Nama dan rekening bank tidak sesuai atau rekening diblokir",
	},
	"SWITCHING_NETWORK_ERROR": LangItem{
		LangEnUS: "SWITCHING_NETWORK_ERROR",
		LangIdID: " Swiching network bank bermasalah, silakan lakukan pembayaran ulang setelah 1-3 jam",
	},
	"TEMPORARY_BANK_NETWORK_ERROR": LangItem{
		LangEnUS: "TEMPORARY_BANK_NETWORK_ERROR",
		LangIdID: " Bank bermasalah sementara dalam pembayaran, silakan lakukan pembayaran ulang setelah 1-2 jam",
	},
	"Transfer Inquiry Decline": LangItem{
		LangEnUS: "Transfer Inquiry Decline",
		LangIdID: "Nama dan rekening bank tidak sesuai atau rekening diblokir",
	},
	"Transfer Decline": LangItem{
		LangEnUS: "Transfer Decline",
		LangIdID: "Pendanaan ditolak oleh bank",
	},
	"置为失效": LangItem{
		LangEnUS: "Set to invalid",
		LangIdID: "kedaluwarsa",
	},
	"记录": LangItem{
		LangEnUS: "record",
		LangIdID: "riwayat",
	},
	"电核情况": LangItem{
		LangEnUS: "Phone verify situation",
		LangIdID: "situasi verifikasi via telepon",
	},
	"电核结果": LangItem{
		LangEnUS: "Phone verify result",
		LangIdID: "hasil verifikasi via telepon",
	},
	"选择拨打时间": LangItem{
		LangEnUS: "Select time to dial",
		LangIdID: "pilih waktu untuk menelepon",
	},
	"仅保存通话记录": LangItem{
		LangEnUS: "Save call history only",
		LangIdID: "menyimpan riwayat telepon saja",
	},
	"保存电核通话记录成功": LangItem{
		LangEnUS: "Save phone verify call history success",
		LangIdID: "Simpan catatan panggilan radio dengan sukses",
	},
	"保存电核通话记录失败": LangItem{
		LangEnUS: "Save phone verify call history only fail",
		LangIdID: "Menyimpan catatan panggilan radio gagal",
	},
	"订单应还款金额": LangItem{
		LangEnUS: "The amount of the order payable",
		LangIdID: " jumlah wajib kembal",
	},
	"订单状态": LangItem{
		LangEnUS: "The order status",
		LangIdID: "Status kembalian ",
	},
	"凭证查看": LangItem{
		LangEnUS: "Credentials to view",
		LangIdID: " Lihat  bukti transfer",
	},
	"凭证": LangItem{
		LangEnUS: "Credentials",
		LangIdID: "bukti transfer",
	},
	"反馈时间": LangItem{
		LangEnUS: "Feedback time",
		LangIdID: "Waktu",
	},
	"还款问题反馈列表": LangItem{
		LangEnUS: "Feedback list of repayment problems",
		LangIdID: "Umpan balik masalah pembayaran",
	},
	"凭证列表": LangItem{
		LangEnUS: "Voucher list",
		LangIdID: "Daftar voucher",
	},
	"优惠券": LangItem{
		LangIdID: "kupon",
	},
	"虚拟还款": LangItem{
		LangIdID: "pembayaran virtual",
	},
	"余额": LangItem{
		LangIdID: "saldo",
	},
	"砍头收取": LangItem{
		LangIdID: "bunga di muka",
	},
	"还款来源": LangItem{
		LangIdID: "sumber pembayaran",
	},
	"退款入账": LangItem{
		LangIdID: "kredit(refund)",
	},
	"退款出账": LangItem{
		LangIdID: "debit(refund)",
	},
	"客户端版本号": LangItem{
		LangEnUS: "Version Name",
		LangIdID: "Version Name",
	},
	"GP中的版本号": LangItem{
		LangEnUS: "Version Code",
		LangIdID: "Version Code",
	},
	"“置为失效”按钮适用情况": LangItem{
		LangEnUS: "'Disabled' button applies",
		LangIdID: "Adegan untuk tombol 'kerdaluwasa'",
	},
	"证件照和手持证件照照片模糊": LangItem{
		LangEnUS: "Photo of passport photo and hand-held ID photo are blurry",
		LangIdID: "Fotokopi foto ID dan foto ID genggam",
	},
	"手持证件照中无人脸或者身份证": LangItem{
		LangEnUS: "Hand-held ID photo with no face or ID card",
		LangIdID: "Foto ID genggam tanpa wajah atau kartu identitas",
	},
	"还款码查询": LangItem{
		LangEnUS: "Paymentcode",
		LangIdID: "Permintaan kode pembayaran",
	},
	"失效时间": LangItem{
		LangEnUS: "Invalid Time",
		LangIdID: "Waktu kegagalan",
	},
	"渠道": LangItem{
		LangEnUS: "Channel",
		LangIdID: "Channel",
	},
	"应还金额": LangItem{
		LangEnUS: "Amount",
		LangIdID: "Jumlah terutang",
	},
	"身份证照片模糊,身份证信息无法识别": LangItem{
		LangEnUS: "ID card photo is blurred, ID card information is not recognized",
		LangIdID: " Foto KTP kabur sehingga informasi pada KTP tidak dapat teridentifikasi",
	},
	"手持证件照模糊": LangItem{
		LangEnUS: "ID hold photo blurred",
		LangIdID: "Foto KTP yang dipegang kabur",
	},
	"手持证件照中缺少人脸": LangItem{
		LangEnUS: "ID hold photo no face",
		LangIdID: "Wajah Anda pada KTP yang dipegang tidak dapat terdeteksi",
	},
	"手持证件照中缺少身份证": LangItem{
		LangEnUS: "ID hold photo no identify card",
		LangIdID: "Kartu identitas yang dipegang bukan KTP",
	},
	"请选择失效原因": LangItem{
		LangEnUS: "Please select invalid reason",
		LangIdID: "Silakan pilih alasan kegagalan",
	},
	"失效原因": LangItem{
		LangEnUS: "Invalid reason",
		LangIdID: "Alasan kegagalan",
	},
	"自动外呼结果": LangItem{
		LangEnUS: "Auto call result",
		LangIdID: "Hasil panggilan otomatis",
	},
	"注册成功": LangItem{
		LangIdID: "Pendaftaran berhasil",
	},
	"登录成功": LangItem{
		LangIdID: "Login berhasil",
	},
	"申请成功": LangItem{
		LangIdID: "Pengajuan pinjaman berhasil",
	},
	"还款成功": LangItem{
		LangIdID: "Pengembalian pinjaman berhasil",
	},
}
