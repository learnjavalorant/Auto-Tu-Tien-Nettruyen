# Auto-Tu-Tien-Nettruyen
Tool Dùng để auto tu tiên tại nettruyen. chỉ dùng để kham thảo.
# Đọc Kĩ
```
1. Tool được viết từ golang (tôi viết chơi chơi nên hơi simple.)
2. Compile để dùng hoặc download từ release.
3. để compile ae install golang https://go.dev/doc/install
4. khi install ae "go build" để compile.
5. để chạy ae click vào là dc.
```
```
git clone https://github.com/learnjavalorant/Auto-Tu-Tien-Nettruyen.git
cd Auto-Tu-Tien-Nettruyen
go build
./AutoTuTien.exe
```
```
Note: đang support mỗi trình duyệt brave bạn có thể thay đổi trong source code.
```

Config
```
ae muốn auto 3 tab thì thêm 1 cái nữa thì làm như này
    "Truyen": [
      {
        "TruyenUrl": "https://www.nettruyenvv.com/truyen-tranh/vo-luyen-dinh-phong-176960",
        "SoChapDoc": 5,
        "TenTruyen": "Võ Luyện Đỉnh Phong 1"
      },
      {
        "TruyenUrl": "https://www.nettruyenvv.com/truyen-tranh/vo-luyen-dinh-phong-176960",
        "SoChapDoc": 5,
        "TenTruyen": "Võ Luyện Đỉnh Phong 2"
      },
      {
        "TruyenUrl": "https://www.nettruyenvv.com/truyen-tranh/vo-luyen-dinh-phong-176960",
        "SoChapDoc": 5,
        "TenTruyen": "Võ Luyện Đỉnh Phong 3"
      }
    ],
    "DelayTime": 1000
  } 
```
