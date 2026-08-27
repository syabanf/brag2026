export type TourStep = {
  id: string;
  title: string;
  body: string;
  /** Narration sent to text-to-speech. Kept separate so it can read naturally. */
  voice: string;
  route: string;
  /** Optional element to scroll to and outline once the route has settled. */
  selector?: string;
};

export const TOUR_STEPS: TourStep[] = [
  {
    id: "welcome",
    title: "Selamat datang di BRAG 2026",
    body: "BRAG adalah platform gamifikasi untuk BNI Grow Annual Challenge. Satu musim berjalan 12 minggu, dan setiap kontribusi anggota menambah poin untuk timnya.",
    voice: "Selamat datang di BRAG 2026. BRAG adalah platform gamifikasi untuk BNI Grow Annual Challenge. Satu musim berjalan dua belas minggu, dan setiap kontribusi anggota menambah poin untuk timnya.",
    route: "/"
  },
  {
    id: "dashboard",
    title: "Dashboard anggota",
    body: "Di sini anggota melihat skor pribadi, posisi timnya, dan booster yang sedang aktif. Semua angka diambil dari score ledger — satu-satunya sumber kebenaran perhitungan poin.",
    voice: "Ini dashboard anggota. Di sini anggota melihat skor pribadi, posisi timnya, dan booster yang sedang aktif. Semua angka diambil dari score ledger, satu-satunya sumber kebenaran perhitungan poin.",
    route: "/"
  },
  {
    id: "leaderboard",
    title: "Leaderboard tim",
    body: "Sepuluh tim diurutkan berdasarkan total poin. Tap salah satu tim untuk melihat rincian riwayat TYFCB dan visitor yang menyusun skornya.",
    voice: "Ini leaderboard tim. Sepuluh tim diurutkan berdasarkan total poin. Tap salah satu tim untuk melihat rincian riwayat TYFCB dan visitor yang menyusun skornya.",
    route: "/leaderboard"
  },
  {
    id: "submit",
    title: "Catat kontribusi",
    body: "Anggota mencatat dua jenis kontribusi: TYFCB dan visitor. Setiap pengajuan masuk berstatus pending sampai diverifikasi admin.",
    voice: "Di halaman ini anggota mencatat kontribusi. Ada dua jenis: TYFCB dan visitor. Setiap pengajuan masuk berstatus pending sampai diverifikasi admin.",
    route: "/submit"
  },
  {
    id: "scoring",
    title: "Cara poin dihitung",
    body: "Poin TYFCB dihitung dari Band dikali Pair Penalty dikali Event Multiplier. Band naik bertingkat dari 10 poin untuk nilai di bawah 500 ribu, sampai 200 poin untuk 500 juta ke atas.",
    voice: "Bagaimana poin dihitung? Poin TYFCB dihitung dari Band, dikali Pair Penalty, dikali Event Multiplier. Band naik bertingkat dari sepuluh poin untuk nilai di bawah lima ratus ribu, sampai dua ratus poin untuk lima ratus juta ke atas.",
    route: "/booster"
  },
  {
    id: "booster",
    title: "Booster & event mingguan",
    body: "Setiap minggu punya satu event aktif yang mengubah pengali poin. Booster seperti Double TYFCB Week bisa melipatgandakan perolehan tim dalam sepekan.",
    voice: "Ini halaman booster dan event mingguan. Setiap minggu punya satu event aktif yang mengubah pengali poin. Booster seperti Double TYFCB Week bisa melipatgandakan perolehan tim dalam sepekan.",
    route: "/booster"
  },
  {
    id: "admin",
    title: "Verifikasi admin",
    body: "Growth Coordinator memverifikasi setiap TYFCB. Saat disetujui, sistem menghitung skor dan menulisnya ke ledger secara permanen — tidak bisa diubah, hanya dikoreksi.",
    voice: "Ini sisi admin. Growth Coordinator memverifikasi setiap TYFCB. Saat disetujui, sistem menghitung skor dan menulisnya ke ledger secara permanen. Tidak bisa diubah, hanya dikoreksi.",
    route: "/admin/tyfcb"
  },
  {
    id: "members",
    title: "Kelola anggota & tim",
    body: "Admin mengatur 100 anggota di 10 tim, menetapkan klasifikasi bisnis, status warna, dan kapten tim.",
    voice: "Admin juga mengelola anggota dan tim. Ada seratus anggota di sepuluh tim, lengkap dengan klasifikasi bisnis, status warna, dan kapten tim.",
    route: "/admin/members"
  },
  {
    id: "awards",
    title: "Badge & penghargaan",
    body: "Dua belas badge diberikan otomatis saat anggota mencapai milestone tertentu, dari First Blood sampai Centurion.",
    voice: "Terakhir, badge dan penghargaan. Dua belas badge diberikan otomatis saat anggota mencapai milestone tertentu, dari First Blood sampai Centurion.",
    route: "/awards"
  },
  {
    id: "done",
    title: "Tur selesai",
    body: "Itu alur lengkap BRAG 2026. Gunakan pemilih peran di pojok kanan atas untuk berpindah antara Admin, Captain, dan Member.",
    voice: "Itu alur lengkap BRAG 2026. Gunakan pemilih peran di pojok kanan atas untuk berpindah antara Admin, Captain, dan Member. Selamat mencoba.",
    route: "/"
  }
];
