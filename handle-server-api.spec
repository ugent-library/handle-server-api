Name: handle-server-api
Summary: handle-server-api
License: BSD
Version: 0.1
Release: X
BuildArch: x86_64
BuildRoot: %(mktemp -ud %{_tmppath}/%{name}-%{version}-%{release}-XXXXXX)

# Requires repository epel-release
BuildRequires: golang

Source: %{name}.tar.gz

#disable creation of debug info (because it fails anyway)
%define  debug_package %{nil}

%description
Temporary rest api that directly inserts data into mysql of handle server

%prep
%setup -q -n %{name}

%build
cd $RPM_BUILD_DIR/%{name} &&
go build -o handle-server-api || exit 1

%install
rm -rf %{buildroot}

mkdir -p %{buildroot}/opt/%{name}
mkdir -p %{buildroot}/etc/systemd/system
mkdir -p %{buildroot}/var/log/%{name}
cp $RPM_BUILD_DIR/%{name}/handle-server-api %{buildroot}/opt/%{name}/handle-server-api
cp $RPM_BUILD_DIR/%{name}/etc/systemd/%{name}.service %{buildroot}/etc/systemd/system/

%clean
rm -rf %{buildroot}

%files
%defattr(-,biblio,biblio,-)
%attr(755,root,root) /etc/systemd/system/%{name}.service
/opt/%{name}/
/var/log/%{name}

%doc

%post
systemctl daemon-reload &&
systemctl enable handle-server-api &&
systemctl restart handle-server-api

exit 0

%preun
if [ $1 -eq "0" ] ; then
  systemctl stop handle-server-api
  systemctl disable handle-server-api
fi

exit 0

%changelog
