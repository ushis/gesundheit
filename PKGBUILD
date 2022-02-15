# Maintainer: ushi <ushi@honkgong.info>
pkgname=gesundheit
pkgver=0.0.1
pkgrel=1
epoch=
pkgdesc="Get notifications about unexpected system state from your local Gesundheitsdienst."
arch=('i686' 'x86_64' 'armv7h' 'aarch64')
url="https://github.com/ushis/gesundheit"
license=('MIT')
groups=()
depends=()
makedepends=('go')
checkdepends=()
optdepends=()
provides=()
conflicts=()
replaces=()
backup=()
options=()
install=
changelog=
source=("$pkgname-$pkgver.tar.gz::https://github.com/ushis/gesundheit/archive/${pkgver}.tar.gz")
noextract=()
sha256sums=('c6cc84272c97b4d4dab056044b8ef6fe46a7873f94f2fb4c0e8d8c0df89a180a')
validpgpkeys=()

#prepare() {}

build() {
	cd "$pkgname-$pkgver"
	export CGO_ENABLED="0"
	export GOFLAGS="-buildmode=pie -trimpath -mod=readonly -modcacherw"
	go build -v -o "$pkgname"
}

#check() {}

package() {
	cd "$pkgname-$pkgver"
	install -Dm755 "$pkgname" "$pkgdir/usr/bin/$pkgname"
	install -Dm644 "systemd/$pkgname.sysusers" "$pkgdir/usr/lib/sysusers.d/$pkgname.conf"
	install -Dm644 "systemd/$pkgname.service" "$pkgdir/usr/lib/systemd/system/$pkgname.service"
	install -Dm644 "LICENSE" "$pkgdir/usr/share/licenses/$pkgname/LICENSE"
	install -dm755 "$pkgdir/etc/$pkgname"
}
