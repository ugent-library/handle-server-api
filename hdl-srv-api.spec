Name: hdl-srv-api
Summary: hdl-srv-api
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
go build -o hdl-srv-api || exit 1

%install
rm -rf %{buildroot}

mkdir -p %{buildroot}/opt/%{name}
mkdir -p %{buildroot}/etc/systemd/system
mkdir -p %{buildroot}/var/log/%{name}
cp $RPM_BUILD_DIR/%{name}/hdl-srv-api %{buildroot}/opt/%{name}/hdl-srv-api
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
systemctl enable hdl-srv-api &&
systemctl restart hdl-srv-api

exit 0

%preun
if [ $1 -eq "0" ] ; then
  systemctl stop hdl-srv-api
  systemctl disable hdl-srv-api
fi

exit 0

%changelog